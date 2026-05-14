# CLOSEOUT — Opik reconsidered (Stage 0 observability substrate)

**Phase:** Stage 0 Week 3, observability deliverable 7.
**Outcome:** Substituted. Opik → Jaeger + OTel collector.
**Date:** 2026-05-14.
**Status:** Second plan-deviation in the project's history. Follows the four-part deviation template established by PR #13 (Mirage Layer-3 → Layer-5).

## Structural finding

The Week 3 inline-plan round's dependency-shape discovery pass surfaced that **Opik's self-hosted deployment is a 13-container platform** (MySQL + Redis + ClickHouse + Zookeeper + MinIO + 5 Opik-specific services + Jaeger + OTel collector), not a single observability service comparable in weight to Postgres or LiteLLM. The original Layer 5 deliverable's framing of Opik as a single observability backend was incorrect.

The mismatch is structural, not measurement-driven. No volume of traces would change the container count of the platform; no Stage 0 evaluation pass would surface a "lighter Opik." The discovery pattern is the same one that fired during PR #13 against Mirage's deployment model: **the marketing description of a dependency is not its deployment shape**. See [`docs/solutions/SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md`](../solutions/SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md) — this Opik finding is the second instance recorded in that pattern (Mirage was first; LiteLLM `wolfi-base` was third at PR-A's CI iteration).

## Why not relax-the-gate

Three options were considered before deciding to substitute:

1. **Run Opik in a stripped-down configuration.** Rejected — Opik's docs don't document a minimal-mode deployment; the 13-container set is the supported topology. Operating an unsupported subset is a maintenance burden Galileo doesn't take on for a non-core dependency.

2. **Defer observability past Stage 0.** Rejected — STAGE_0_PLAN.md §Week 3 deliverable 7 names span emission as a Stage 0 gate-test requirement (100 demo runs, 100 parent spans). Removing the deliverable would require relaxing a pre-registered gate, which v7 rule 9 specifically refuses.

3. **Substitute a lighter substrate that satisfies the gate test.** Adopted. Jaeger + OTel collector cover trace ingest + storage + query in two containers (vs Opik's thirteen). The trace shape required by the gate test (parent spans per request, child spans per LLM call) is OTel-native; Opik is itself an OTel consumer downstream.

## What ships in PR-B

| Artifact | Path | Purpose |
|---|---|---|
| Closeout (this file) | `docs/closeouts/CLOSEOUT_OPIK_RECONSIDERED.md` | Names the structural finding; required by v7 rule 3 for any phase or deliverable with a pre-registered gate, pass *or* fail. |
| ADR | `docs/decisions/0004-observability-substrate.md` | Locks in Jaeger + OTel collector; documents reversal triggers if a different substrate becomes warranted in Stage 1+. |
| Plan edits | `docs/plans/STAGE_0_PLAN.md` §Week 3 deliverable 7 | Removes "deferred to PR-B" line; records substitution; preserves the gate-test requirement unchanged. |
| Code | `kernel/cmd/gateway/otel.go`, compose YAML additions, CI service containers, wrapper image | Emits spans through the substituted substrate; same gate-test contract holds. |

All four artifacts land in PR-B (the single-commit-set discipline from PR #13 holds — the deviation is indivisible from the code that implements it).

## Reversal triggers

This substitution should be reconsidered if any of the following fire in Stage 1+:

- An observability stack that Galileo's operators actually want to run (per tenant evaluation feedback) is non-OTel-native and integrates more cleanly with Opik than with Jaeger.
- Gate-test trace volume or query complexity exceeds Jaeger's documented limits.
- A Galileo customer requires the LLM-evaluation surface (prompt diff playgrounds, eval dashboards) that Opik provides natively and Jaeger doesn't.

Until then, two containers handle Stage 0's observability gate. The thirteen-container alternative is the wrong shape for the deliverable.

## Cross-references

- Origin of the discovery: PR-A inline plan, 2026-05-14, captured in [`docs/plans/STAGE_0_PLAN.md`](../plans/STAGE_0_PLAN.md) §Week 3 deliverable 7 (pre-PR-B revision).
- Pattern documentation: [`docs/solutions/SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md`](../solutions/SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md) — Opik is instance 2 of 3.
- Deviation template precedent: [`docs/closeouts/CLOSEOUT_LAYER3_MIRAGE_RECONSIDERED.md`](CLOSEOUT_LAYER3_MIRAGE_RECONSIDERED.md), 2026-05-13.
- ADR for this substitution: [`docs/decisions/0004-observability-substrate.md`](../decisions/0004-observability-substrate.md).
