# ADR-0006 — Stage 1 scope: canonical §6 customer-facing MVP, 6-week target, Path 2 discipline calibration

| Field | Value |
| --- | --- |
| **Status** | Accepted |
| **Date** | 2026-05-16 |
| **Decider** | Emmanuel (founder) |
| **Author** | Claude Opus 4.7 (1M context) under Emmanuel's direction |
| **Type** | **Forward scope decision** — not a plan-deviation closeout. The first ADR that *commits to a scope and a calibration* rather than *retroactively documenting a discovery*. |
| **Supersedes** | The Stage 0 closeout's 13-entry deferral table as the Stage 1 *scope ledger* (the deferral table is preserved as architectural carry-over input but is structurally a subset of Stage 1's canonical scope per [`docs/closeouts/CLOSEOUT_STAGE0.md`](../closeouts/CLOSEOUT_STAGE0.md) §4 + Drift-15) |
| **Companion artifacts** | [`docs/plans/STAGE_1_PLAN.md`](../plans/STAGE_1_PLAN.md) (created in PR-E); the four cosmetic plan-doc edits for Drift-16/-17/-18 in [`docs/galileo_os_infrastructure_plan.md`](../galileo_os_infrastructure_plan.md) (also PR-E); Brain substrate code in `kernel/cmd/gateway/migrations/0005_brain.sql` + `deploy/compose/postgres-brain/` (PR-E) |

## Context

Stage 0 closed at PR #20 with three plan-deviations resolved (Mirage Layer 3→5, Opik→Jaeger, MCP per-source) and a 13-entry deferral table written into the Stage 0 closeout's §4. The deferral table captures Stage 0 → Stage 1 *architectural carry-over*: every Stage 0 simplification that named a structural trigger for Stage 1 reconsideration.

The Stage 1 structural reading round (2026-05-16) surfaced **Drift-15**: the closeout's 13-entry deferral table is structurally **smaller** than the canonical plan §6's Stage 1 scope. Reading them side-by-side:

- **Deferral table:** ~6 architectural components (Brain wiring, Onboarding Crew completion, Skill registry, Operator-review UI, plus 9 reconsideration triggers that aren't deliverables themselves).
- **Canonical plan §6:** ~14 components — five marketing agents, ~40 Skills mirrored + custom, Promptfoo eval, React Native mobile, Telegram bot, Supabase Auth, web admin (4 sub-components), connector expansion (3→8), destructive-action lockdown enforcement, Mirage snapshot-before-write.

Both framings are correct. The deferral table is "what got deliberately deferred from Stage 0"; the canonical plan is "what Stage 1 ships to customers." They diverge because Stage 1's first-customer surface includes work that wasn't in Stage 0's scope to defer.

**Drift-15 is a *scope decision*, not a plan-doc edit.** Three defensible Stage 1 scopes were considered (Stage 1 structural reading round):

- **Scope A** — full canonical §6 customer-facing MVP. ~14 components. Honest 12-14 weeks technical + 4-6 weeks gate satisfaction. Produces customer signal.
- **Scope B** — deferral-table subset only. ~6 components. 6-8 weeks technical. Architectural completion without customer signal.
- **Scope C** — something else entirely (not explored; not warranted by current state).

## Decision

**Stage 1 scope = Scope A (full canonical §6 customer-facing MVP).**

**Technical-surface target: 6 weeks (2026-05-16 → 2026-06-27).** Aggressive compression of the honest 12-14 week projection. Compression is achieved through Path 2 discipline calibration (below).

**Commercial gate target: ~4 weeks after technical surface completes** (5 paying tenants × 2 weeks sustained daily-artifact + 50%-tooling-spend-reduction signal). The commercial half of the gate **cannot be compressed by engineering speed**; it has its own calendar.

**Total Stage 1 closeout target: ~10 weeks from today (~2026-07-25).**

### Path 2 — discipline calibration by component risk

Not every Stage 1 PR earns the full read-first verification + inline plan + diff readback round that Stage 0's complex PRs (PR-A, PR-B, PR-C, PR-D) received. Path 2 calibration recognizes that the architecturally novel components carry the discovery risk; the repeatable-pattern components reuse established patterns with lighter planning rounds.

**Full discipline (~10 components):**

| # | Component | Why full discipline |
|---|---|---|
| [A] | Brain substrate (pgvector + AGE on Postgres) | First custom Dockerfile in the project; first cross-extension Postgres image; first multi-step extension boot sequence |
| [B] | Ingestion Agent (Docling + MarkItDown + Whisper + semchunk) | First Brain consumer; introduces four new heavy Python deps; latency/cost gate target depends on this |
| [C] | Org-Mapper Agent | First LLM-grounded synthesis agent; new shape; §3.5 threshold #3 depends on it |
| [D] | Skill registry + Skill-Selector Agent | First cross-language coupling (Go service + Python agent); §3.5 threshold #4 depends on it |
| [E] | QA Agent | First cross-agent checker pattern; consumes all prior outputs |
| [F] | Operator-review UI | 4 sub-components; replaces CLI invocation surface; first Next.js work beyond the Hello Agent shell |
| [G] | Approval-state machine (Temporal signal gate) | Foundational for every write-scope agent; destructive-action lockdown's enforcement surface |
| [M] | Supabase Auth | Replaces dev JWT keypair; substantial walkthrough rewrite |
| [L] | Destructive-action lockdown enforcement + Mirage snapshot-before-write | First Mirage Layer-5 adopter; load-bearing for the §3.5 destructive=0 threshold |
| [K-first] | Mobile + Telegram approval surface (first PR establishing the pattern) | First React Native + Expo work; new approval-surface contract |

**Streamlined discipline (~38 components):**

- Marketing agents 2-5 (CMO establishes pattern; specialists 2-5 reuse with shorter inline-plan rounds)
- Individual Skills past format adoption (~40 Skills total; format adoption is the novel PR; per-batch lighter rounds)
- Connectors past dispatch (Gmail + X + LinkedIn + GA + Plausible + Stripe — each reuses PR-D's per-source dispatch pattern)
- Promptfoo eval harness (single CI integration; mostly mechanical)
- Walkthrough/protocol doc PRs (Stage 0 hygiene; small)

**Two non-negotiable disciplines even at 6-week pace:**

1. **Dependency-shape verification stays full strength.** Every new external dep gets the four-axis checklist (deployment topology / container topology / ecosystem utilities / distribution channel) per [`docs/solutions/SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md`](../solutions/SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md). Stage 1 deps requiring this: pgvector, AGE, Docling, MarkItDown, insanely-fast-whisper, Trafilatura, semchunk, bge-large-en-v1.5, Promptfoo, React Native, Expo, Telegram Bot API, Supabase Auth.

2. **Plan-deviation handling stays full strength.** Expected cadence per Stage 0 closeout pattern #1: ~1 deviation per architectural week, ~6 deviations across Stage 1. Each gets the four-part template (closeout + ADR + plan edits + code in one indivisible PR). Zero deviations across four Stage 1 weeks indicates discipline-skipping per the Stage 0 closeout's §7 framing, not perfect planning.

## Reversal triggers (the adjustment ladder)

The 6-week commitment is **safe rather than aspirational** because of the named adjustment ladder. If reality intervenes, scope adjusts in a pre-registered order (most-acceptable-cut first):

1. **Week-3 review trigger.** If aggregate PR velocity is < 4 PRs/week or any [A]–[G] component is more than 1 week behind plan: **drop React Native mobile app [K]** (Telegram bot covers the approval surface). Stage 2 carries React Native as a deferred deliverable.
2. **Week-4 review trigger.** If aggregate PR velocity is still < 4 PRs/week: **drop marketing agents 4-5 [H4-H5]** (CMO + 3 specialists covers the daily-artifact gate signal). Stage 2 carries Ad Ops + Growth Analyst.
3. **Week-5 review trigger.** If still behind: **reduce Skills count from ~40 to ~25** (highest-value-first selection — gooseworks GTM core + CosmoBlk email-marketing-bible essentials).
4. **Last-resort trigger.** If [N] Promptfoo eval is itself the bottleneck: **defer [N] entirely to Stage 2** and ship Stage 1 without eval-on-every-Skill-PR. Skill versioning still occurs; eval becomes Stage 2's first PR.

The ladder commits to 6 weeks **with named order-of-cuts**, not 6 weeks at all costs. The trigger fires on velocity signals at the week-N review checkpoints, not on a single bad week.

## Consequences

**Operational.** Stage 1 introduces:
- One custom Dockerfile (`deploy/compose/postgres-brain/`) — first Galileo-built image in the project (second instance of the runtime-introduction size-exception category, cites PR-A).
- ~10 new Python runtime deps across [B] alone (Docling, MarkItDown, Whisper, Trafilatura, semchunk, bge-large embeddings, plus their transitives).
- One new Go service (Skill registry, [D]).
- Four new web admin components (org-chart, ticket queue, calendar, Brain explorer).
- One new React Native app + one new Telegram bot service.
- Two new external auth surfaces (Supabase Auth; the OAuth flows for the five new connector source-kinds).

**Code.** PR-E (this PR) ships [A] Brain substrate as the foundational PR for all of [B]–[E]. Subsequent PRs unblock in the dependency chain per STAGE_1_PLAN.md §3.

**Plan & spec.** STAGE_1_PLAN.md created in this PR mirrors STAGE_0_PLAN.md structure with the 14 components, prerequisite chain, per-component PR projection (Path 2 calibration explicit per component), six-week calendar, and adjustment ladder. STAGE_0_PLAN.md gets a closing pointer-forward note. Canonical plan §3.3 step 4 and §4.6 get cosmetic Drift-16/-17/-18 updates.

**Discipline.** Drift-numbering continues from Drift-14 (the last Stage 0 drift). Drift-15 (this scope decision) lands as this ADR; Drift-16/-17/-18 land as cosmetic canonical-plan edits in this PR. Future Stage 1 drifts continue the sequence from Drift-19.

**Reconsideration triggers** (Stage 1 hygiene, separately from the adjustment ladder):

- Promote to GHCR publishing when ≥3 custom images exist in `deploy/compose/` (currently 2: `otel-wrapper/`, `postgres-brain/`). Until then, build-in-CI per the PR-E CI pattern.
- When AGE 1.7.0 publishes a stable Docker image (currently only `release_PG17_1.6.0` is Dockerized; `PG17/v1.7.0-rc0` GitHub tag exists since 2026-02-11), evaluate upgrade in a separate PR.
