# `kernel/probe/mirage/` — Stage 0 Mirage probe apparatus

This package implements the **measurement apparatus** for the Stage 0 Mirage probe specified in [`docs/plans/STAGE_0_PLAN.md`](../../../docs/plans/STAGE_0_PLAN.md) §Week 2. It does **not** import Mirage.

## What's here (plan-PR #10)

- `apparatus.go` — the `Workspace` interface + the five probe functions (`RunOAuthProbe`, `RunCacheProbe`, `RunSnapshotProbe`, `RunListProbe`, `RunStatProbe`) + result types that name specific failure modes rather than reporting generic errors.
- `mocks_test.go` — synthetic mock `Workspace` implementations (happy + per-failure-mode). Test-only; never shipped in production.
- `apparatus_test.go` — apparatus self-validation (manifest hash determinism, seed uniqueness).
- `{oauth,cache,snapshot,list,stat}_test.go` — five probe tests with two-or-more variants each: a happy-path mock that the apparatus must report `Pass=true` against, plus failure-injection mocks that the apparatus must catch as the specific failure mode each injects.

## What's NOT here (plan-PR #11)

- `mirage_backend.go` — wraps the vendored `strukto-ai/mirage` in the `Workspace` interface.
- `mcp-servers/mirage-vendored/` — pinned-commit vendor.
- `integration_test.go` — runs the five probes against the live Mirage backend under `//go:build integration`.
- `docs/closeouts/PROBE_MIRAGE_STAGE0.md` — pass/fail closeout with structural finding and downstream architectural consequences.

## Why apparatus before implementation

v7 rule 4 (calibration artifacts before implementation). If the apparatus and Mirage shipped together and the probe failed, "did the apparatus fail or did Mirage fail?" would be unanswerable. By the time Mirage is measured, the apparatus has been independently validated against synthetic happy-path mocks AND named failure-injection mocks. The probe's outcome is therefore a measurement of Mirage, not of the measurement.

If the apparatus fails its own validation (happy-path mock makes the apparatus report failure, or a failure-injection mock makes the apparatus report pass), the apparatus is invalid. Per v7 rule 9, the harness is redesigned and `docs/closeouts/CLOSEOUT_PROBE_APPARATUS.md` names the structural finding. Mirage adoption blocks until self-validation passes.

## Running

```bash
make probe            # from repo root
# or
cd kernel && go test -count=1 -v ./probe/mirage/...
```

Wall-clock budget: all six happy-path tests + seven failure-injection tests complete in well under 30s on a CI runner; the synthetic-input scale is chosen so mocks respond in memory.

Seeds for any randomized test are logged via `t.Logf("seed=...")` so a flaky failure is reproducible.
