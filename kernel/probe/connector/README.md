# `kernel/probe/connector/` — Workspace connector verification harness

This package implements a **general kernel-side verification apparatus** for any data backend that implements the `Workspace` interface. The apparatus measures whether a Workspace implementation satisfies named failure-mode criteria (cross-tenant isolation, cache p99 latency, content corruption, content staleness, byte-identical snapshot/restore, dropped-file detection, listing accuracy, stat freshness).

The apparatus is independently validated against synthetic mock backends in `*_test.go` files in this package. It does **not** import or depend on any specific backend.

## History

Originally written as the Stage 0 Mirage probe apparatus (PR #10, `kernel/probe/mirage/`). After PR #13 relocated Mirage to Layer 5 (agent-side) per ADR-0003, the apparatus was renamed and retained as a general kernel-side connector probe. Future Layer 3 substrate candidates (S3, Postgres, discrete MCP wrappers, future Mirage if a Go SDK or native server ships per ADR-0003 reversal triggers) implement `Workspace` and are measured by this apparatus.

## What's here

- `apparatus.go` — the `Workspace` interface + five probe functions (`RunOAuthProbe`, `RunCacheProbe`, `RunSnapshotProbe`, `RunListProbe`, `RunStatProbe`) + result types that name specific failure modes rather than reporting generic errors.
- `mocks_test.go` — synthetic mock `Workspace` implementations (happy + 9 named-failure-mode + 4 error-propagation). Test-only; never shipped in production.
- `apparatus_test.go` — apparatus self-validation (manifest hash determinism, seed uniqueness).
- `{oauth,cache,snapshot,list,stat}_test.go` — five probe tests with two-or-more variants each: a happy-path mock that the apparatus must report `Pass=true` against, plus failure-injection mocks that the apparatus must catch as the specific failure mode each injects.
- `errorpaths_test.go` — coverage drill-down tests for Workspace-method error propagation, defensive input validation, and missing-failure-mode diagnostics (see `docs/solutions/SOLUTION_COVERAGE_DRILL_DOWN.md`).

## Why apparatus before any candidate is graded

v7 rule 4 (calibration artifacts before implementation). If the apparatus and a candidate backend shipped together and the probe failed, "did the apparatus fail or did the candidate fail?" would be unanswerable. The apparatus is independently validated against synthetic happy-path mocks AND named failure-injection mocks. Any probe outcome is therefore a measurement of the candidate, not of the measurement.

If the apparatus fails its own validation (happy-path mock makes the apparatus report failure, or a failure-injection mock makes the apparatus report pass), the apparatus is invalid. Per v7 rule 9, the harness is redesigned and a `CLOSEOUT_PROBE_APPARATUS.md` names the structural finding. Adoption of any candidate blocks until self-validation passes.

## Running

```bash
make probe            # from repo root
# or
cd kernel && go test -count=1 -v ./probe/connector/...
```

Wall-clock budget: all 26 tests complete in ~1.2s on a CI runner against a 30s budget; the synthetic-input scale is chosen so mocks respond in memory. Coverage is 98.1% (the only uncovered branch is `freshSeed`'s `crypto/rand` fallback, documented as Bucket 3 in `docs/solutions/SOLUTION_COVERAGE_DRILL_DOWN.md`).

Seeds for any randomized test are logged via `t.Logf("seed=...")` so a flaky failure is reproducible.
