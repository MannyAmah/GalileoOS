# ADR-0003 — Mirage placement: Layer 5 (agent-side library) instead of Layer 3 (kernel-side substrate)

| Field | Value |
| --- | --- |
| **Status** | Accepted |
| **Date** | 2026-05-13 |
| **Decider** | Emmanuel (founder) |
| **Author** | Claude Opus 4.7 (1M context) under Emmanuel's direction |
| **Supersedes** | The Layer 3 placement of Mirage in [`docs/galileo_os_infrastructure_plan.md`](../galileo_os_infrastructure_plan.md) §4.4 (table row removed) and the Stage 0 Mirage probe in [`docs/plans/STAGE_0_PLAN.md`](../plans/STAGE_0_PLAN.md) §Week 2 (revised to closeout delivery) |
| **Plan deviation** | Yes — **first plan-deviation in the project's history.** The shape used (closeout + canonical plan edits + this ADR + follow-up code-rename PR) is the template for future deviations. |
| **Companion artifacts** | [`docs/closeouts/CLOSEOUT_LAYER3_MIRAGE_RECONSIDERED.md`](../closeouts/CLOSEOUT_LAYER3_MIRAGE_RECONSIDERED.md) — full structural finding and reasoning chain. This ADR carries the metadata and reversal machinery only. |

## Context

The original plan placed Mirage at Layer 3 (Brain substrate) with a Stage 0 probe gating adoption. During PR #13's inline planning round (2026-05-13), reading Mirage's deployment-model documentation directly surfaced the structural mismatch driving this ADR: Mirage ships Python and TypeScript SDKs only, with no Go SDK and no native server. Mirage is designed for in-process embedding inside agent code (FastAPI, Express, browser apps, async runtimes). Galileo's kernel is Go.

Three readings were considered (full reasoning in the companion closeout):

1. **Layer 3 with permanent Python sidecar.** Rejected — heavy operational commitment for a v0.0.1 dependency; forfeits the "one service = one language" discipline.
2. **Layer 5 as agent-side library.** **Chosen.** Fits Mirage's own deployment model; preserves kernel-language purity; per-agent choice.
3. **Drop Mirage entirely.** Rejected — Mirage's unified-filesystem value proposition is real at the agent layer.

The closeout documents four reasons Reading 2 beats Reading 1: language-purity alignment, no permanent operational cost, clean shift of destructive-action defenses from kernel to agent code, and structural verifiability without a live probe.

## Decision

Mirage is placed at **Layer 5 (Integrations and Tools), agent-side**. Concretely:

- Galileo's Go kernel does **not** import Mirage. No Go SDK exists; no sidecar is built or run.
- Python agents that want Mirage's unified-filesystem abstraction `import mirage-ai` directly and use it in-process.
- TypeScript agents (future) use `@struktoai/mirage-*` similarly.
- Per-agent choice between Mirage and direct connector MCP clients. Both coexist in the same Onboarding Crew.
- The destructive-action lockdown's pre-write snapshot requirement is enforced kernel-side via Temporal signal; the snapshot **artifact** is produced agent-side (Mirage's `workspace.snapshot()` for Mirage-using Python agents; source-native mechanisms or Galileo-produced backups otherwise). Postgres PITR covers Brain-state durability as a separate defense.
- PR #10's `Workspace`-interface verification harness is retained as a general kernel-side connector probe (rename to a more accurate name in a follow-up PR; name decided with code context).

## Consequences

**Positive:**

- Kernel-language discipline preserved (Go-only kernel; no Python sidecar in production).
- No permanent operational cost for a v0.0.x dependency.
- Mirage's value proposition is preserved at the layer where it fits.
- The destructive-action defense responsibility splits cleanly: kernel enforces *existence* of artifact, agents produce it.
- PR #10's apparatus generalizes rather than gets discarded.

**Negative / risks:**

- Future Galileo kernel needs (e.g., kernel-level crawling of heterogeneous backends) cannot use Mirage. If such a need arises, the trigger (c) below fires and the placement is re-evaluated.
- Per-agent choice introduces manifest schema variability (Mirage-using agents and Mirage-bypassing agents). Mitigation: schema designed to be union of both modes; `source_kind` field discriminates (already documented in STAGE_0_PLAN.md §Week 4 risks).
- v0.0.x API breakage risk now scopes to agent code (`agents/*/`) rather than platform-wide. Integration tests pin `mirage-ai==<version>`; if a breaking change doesn't justify the patch cost, affected agents fall back to discrete connector clients.

## Triggers to revisit (reverse or revise this ADR)

This ADR should be **reversed or revised** when **any** of the following happens:

(a) **Mirage publishes an official Go SDK** with stability commitments matching Galileo's kernel needs. Reversal scope: kernel may import Mirage at Layer 3 directly; Layer 5 agent-side use continues coexisting.
(b) **Mirage publishes a native server mode** exposing its API over a documented network protocol (HTTP, gRPC), maintained upstream — not a sidecar Galileo writes itself. Reversal scope: kernel may talk to Mirage as a network service; sidecar burden is upstream's, not Galileo's.
(c) **Galileo's kernel acquires a need to crawl heterogeneous backends itself** that cannot be delegated to Python agents. This is a Galileo-side architectural change more than a Mirage-side change; the placement would be re-evaluated against then-current Mirage state.

Reversal procedure (not a one-line API call like ADR-0002; this is a structural reversal):

1. New closeout doc (`docs/closeouts/CLOSEOUT_LAYER_RELOCATION_REVERSED.md` or equivalent) names which trigger fired, what changed upstream, and the new placement decision.
2. Canonical plan edits restore Layer 3 references (or revise to whatever the new placement is).
3. New ADR (`docs/decisions/000N-mirage-layer-revision.md`) supersedes this one.
4. Code changes follow per the new placement (kernel-side imports if (a); kernel-side network client if (b); kernel-side crawler if (c)).

This mirrors the four-part deviation pattern this ADR established (closeout + plan edits + ADR + code change), applied in reverse.

## References

- [`docs/closeouts/CLOSEOUT_LAYER3_MIRAGE_RECONSIDERED.md`](../closeouts/CLOSEOUT_LAYER3_MIRAGE_RECONSIDERED.md) — full structural finding, three readings, four reasons, maker-checker provenance, template-for-future-deviations.
- [`docs/galileo_os_infrastructure_plan.md`](../galileo_os_infrastructure_plan.md) §4.4 (Layer 3 note on Mirage's placement), §4.6 (Layer 5 agent-side row), §3.3 step 4 (Mount), Risk Register (Destructive-action row, Mirage-API-breakage row), §9 Build-vs-Buy.
- [`docs/plans/STAGE_0_PLAN.md`](../plans/STAGE_0_PLAN.md) §Week 2 — revised scope.
- [`CLAUDE.md`](../../CLAUDE.md) §Escalation — plan-deviation discipline ("do not silently work around plan defects"); this ADR is the canonical example of that discipline executed.
- Mirage upstream documentation (`mirage-ai` on PyPI; `@struktoai/mirage-*` on npm) — "Python and TypeScript SDKs give your AI agents a virtual filesystem directly inside FastAPI, Express, browser apps, or any async runtime, no separate process required." This is the source quote that surfaced the language mismatch.
- [`docs/decisions/0001-repo-namespace.md`](./0001-repo-namespace.md), [`docs/decisions/0002-protection-approval-relaxation.md`](./0002-protection-approval-relaxation.md) — prior ADRs (structural pattern).
