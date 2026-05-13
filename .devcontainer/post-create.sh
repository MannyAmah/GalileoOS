#!/usr/bin/env bash
#
# Post-create hook for the Galileo OS devcontainer.
#
# Installs the exact tool versions pinned by CI so `make ci-local` runs
# identically in the devcontainer and on GitHub Actions runners. Every
# version here is matched against .github/workflows/ci.yml; the
# co-change policy in CLAUDE.md governs bumps.

set -euo pipefail

PIN_BUF="1.45.0"
PIN_RUFF="0.6.*"
PIN_MYPY="1.11.*"
PIN_BLACK="26.*"
PIN_PIP_AUDIT="2.7.*"
PIN_GOLANGCI_LINT="v2.12.2"
PIN_GOVULNCHECK="v1.1.4"

echo "==> Python tools (ruff, mypy, black, pip-audit) at CI pins"
pip install --user --no-cache-dir \
  "ruff==${PIN_RUFF}" \
  "mypy==${PIN_MYPY}" \
  "black==${PIN_BLACK}" \
  "pip-audit==${PIN_PIP_AUDIT}"

echo "==> buf at ${PIN_BUF}"
go install "github.com/bufbuild/buf/cmd/buf@v${PIN_BUF}"

echo "==> golangci-lint at ${PIN_GOLANGCI_LINT}"
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
  | sh -s -- -b "$(go env GOPATH)/bin" "${PIN_GOLANGCI_LINT}"

echo "==> govulncheck at ${PIN_GOVULNCHECK}"
go install "golang.org/x/vuln/cmd/govulncheck@${PIN_GOVULNCHECK}"

echo "==> web/ npm install"
(cd web && npm install --no-audit --no-fund)

echo "==> devcontainer ready. Run 'make ci-local' to verify."
