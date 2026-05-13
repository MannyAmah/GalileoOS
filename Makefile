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
up:  ## Stage 0 placeholder — Appendix B docker-compose stack lands in Week 3.
	@echo "make up: not yet wired."
	@echo "Stage 0 Week 3 lands the docker-compose stack from docs/galileo_os_infrastructure_plan.md Appendix B."

.PHONY: probe
probe:  ## Stage 0 placeholder — Mirage probe harness lands in plan-PR #10.
	@echo "make probe: not yet wired."
	@echo "The Mirage probe harness lands in plan-PR #10 (immediately after plan-PR #9 merges)."
	@echo "Week 2 uses this target to run the three probe tests."
