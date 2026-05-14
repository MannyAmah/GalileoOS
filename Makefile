# Galileo OS — Stage 0 Makefile
#
# Bare-metal compatible: each target works if local tool versions match
# the pins in .devcontainer/post-create.sh. The devcontainer is the easy
# path, not the only path. See CLAUDE.md "Tool version pins" for the
# co-change policy between CI and devcontainer.

.DEFAULT_GOAL := help

.PHONY: help
help:  ## Show this help.
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_-]+:.*##/ { printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: ci-local
ci-local:  ## Reproduce the exact CI matrix locally (matches .github/workflows/ci.yml).
	@echo "==> codegen consistency (make generate must produce no diff)"
	@cd schemas && buf generate
	@git diff --exit-code kernel/gen/ || (echo "kernel/gen/ has diff: run 'make generate' and commit the result" && exit 1)
	@echo "==> kernel/  (go vet + build + test + lint + vuln)"
	@cd kernel && go vet ./... && go build ./... && go test ./...
	@cd kernel && golangci-lint run --timeout=3m ./...
	@cd kernel && govulncheck ./...
	@echo "==> agents/  (ruff + black --check + mypy --strict + pip-audit)"
	@cd agents && ruff check . && black --check . && mypy --strict onboarding hello && pip-audit --strict
	@echo "==> web/     (tsc --noEmit + eslint + npm audit)"
	@cd web && npx --no-install tsc --noEmit && npx --no-install eslint . && npm audit --omit=dev --audit-level=high
	@echo "==> schemas/ (buf lint + breaking)"
	@cd schemas && buf lint && buf breaking --against '../.git#ref=refs/remotes/origin/main,subdir=schemas'
	@echo "==> ci-local: all jobs passed"

.PHONY: lint
lint:  ## Run only the lint subset of ci-local.
	@cd kernel && go vet ./... && golangci-lint run --timeout=3m ./...
	@cd agents && ruff check . && black --check .
	@cd web && npx --no-install eslint .
	@cd schemas && buf lint

.PHONY: test
test:  ## Run only the test subset of ci-local.
	@cd kernel && go test ./...

.PHONY: up
up:  ## Bring up the Stage 0 compose stack (postgres + temporal + litellm).
	@docker compose -f deploy/compose/docker-compose.yml up -d
	@echo "==> compose up. Wait for healthchecks: docker compose -f deploy/compose/docker-compose.yml ps"

.PHONY: down
down:  ## Tear down the Stage 0 compose stack and clear the postgres volume.
	@docker compose -f deploy/compose/docker-compose.yml down -v

.PHONY: stage0-jwt-setup
stage0-jwt-setup:  ## Generate the local Ed25519 dev keypair at kernel/auth/dev-keys/ (gitignored).
	@cd kernel && go run ./cmd/jwt-tool genkey -dir auth/dev-keys

.PHONY: stage0-jwt
stage0-jwt:  ## Mint a dev JWT. Usage: make stage0-jwt TENANT=<uuid> [TTL=1h] [BUDGET=50000]
	@if [ -z "$(TENANT)" ]; then echo "TENANT=<uuid> required"; exit 2; fi
	@cd kernel && go run ./cmd/jwt-tool mint \
		-priv auth/dev-keys/private.pem \
		-tenant "$(TENANT)" \
		-ttl "$${TTL:-1h}" \
		-budget "$${BUDGET:-0}"

.PHONY: stage0-gateway-test
stage0-gateway-test:  ## Run the gateway integration suite. Requires `make up` first.
	@cd kernel && go test -tags=gateway_integration -count=1 -v ./cmd/gateway/...

.PHONY: stage0-agent-runner-test
stage0-agent-runner-test:  ## Run the agent-runner integration suite (includes E2E). Requires `make up` + a running gateway subprocess + LiteLLM mock or provider credentials.
	@cd kernel && GALILEO_AGENT_RUN_E2E=1 go test -tags=agent_runner_integration -count=1 -v ./cmd/agent-runner/...

.PHONY: generate
generate:  ## Regenerate protobuf Go stubs (developer step; CI verifies the tree matches schemas).
	@cd schemas && buf generate
	@echo "==> generated kernel/gen/galileo/v1/*.pb.go. Commit alongside the .proto change."

.PHONY: probe
probe:  ## Run the Workspace connector probe apparatus tests (synthetic mocks; no real backend import).
	@cd kernel && go test -count=1 -v ./probe/connector/...
