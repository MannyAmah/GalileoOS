# `schemas/` — Protobuf inter-service contracts

This directory holds the protobuf wire-format contracts that every Galileo service speaks. CI runs `buf lint` and `buf breaking` on every PR against `main`.

## Layout

```
schemas/
  buf.yaml         buf module config + lint rules (STANDARD)
  buf.gen.yaml     codegen config (wired in plan-PR #10+, empty in Stage 0)
  galileo/v1/      versioned protobuf package — every message is in galileo.v1
    tenant.proto       TenantId, TenantContext
    agent_task.proto   TaskInput, SkillRef, AgentOutput, ToolCall, TaskResult
    brain.proto        BrainQuery, BrainResponse, DocumentRef
```

Breaking changes go in `galileo/v2/` rather than mutating `v1`. The `go_package` option on each file routes Go codegen to `kernel/gen/galileo/v1/`.

## Why `buf breaking --against` uses `ref=refs/remotes/origin/main`

The CI workflow runs:

```yaml
buf breaking --against '../.git#ref=refs/remotes/origin/main,subdir=schemas'
```

This is **intentional**. The natural-looking alternative — `--against '../.git#branch=main,subdir=schemas'` — does not work on GitHub Actions PR runners. Here's why:

- `--against '.git#branch=main'` resolves to `refs/heads/main` (the **local** branch).
- `actions/checkout` on PR runners populates **only** the remote-tracking ref (`refs/remotes/origin/main`). A local `main` branch is never created.
- Result: `buf breaking` fails with `fatal: couldn't find remote ref main` and the protobuf job exits 1.

Using `ref=refs/remotes/origin/main` resolves directly against the remote-tracking ref that `actions/checkout` does populate, so `buf breaking` finds the comparison target every time.

Future contributor who tries to "simplify" this to `branch=main`: don't. The simpler-looking syntax breaks CI on every PR. See [`docs/solutions/SOLUTION_CI_GUARD_VERIFICATION.md`](../docs/solutions/SOLUTION_CI_GUARD_VERIFICATION.md) for the full defect chain (PR #5 → #6 → #8 → #7) that produced this constraint.

## Running buf locally

```bash
# from repo root
cd schemas
buf lint
buf breaking --against '../.git#ref=refs/remotes/origin/main,subdir=schemas'
```

Or use the unified entry point that matches CI exactly:

```bash
make ci-local   # runs all four CI jobs locally
```

The local `buf` version must match the CI pin (`1.45.0`). Devcontainer ships this version; bare-metal users should install via `go install github.com/bufbuild/buf/cmd/buf@v1.45.0`.
