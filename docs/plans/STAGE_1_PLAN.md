# Stage 1 Execution Plan — Galileo for Marketing (MVP, weeks 5–10)

| Field | Value |
| --- | --- |
| **Authoritative scope** | [`docs/galileo_os_infrastructure_plan.md`](../galileo_os_infrastructure_plan.md) §6 / "What ships at the end of Stage 1" |
| **Scope decision** | [`docs/decisions/0006-stage1-scope-canonical-with-path2-calibration.md`](../decisions/0006-stage1-scope-canonical-with-path2-calibration.md) (Drift-15 resolution; Scope A approved 2026-05-16) |
| **Phase window** | 2026-05-16 → 2026-06-27 technical surface (6 weeks); commercial gate satisfaction ~2026-07-25 (10 weeks total) |
| **Predecessor phase** | Stage 0 closed at PR #20 (2026-05-16). Stage 0 closeout: [`docs/closeouts/CLOSEOUT_STAGE0.md`](../closeouts/CLOSEOUT_STAGE0.md) |
| **Phase exit** | Stage 1 gate = 5 paying tenants × 2 weeks daily-artifact + 50%-tooling-spend-reduction signal, AND §3.5 perf gate satisfied for each tenant |

## 0. Why this document exists

STAGE_0_PLAN.md was the live execution plan for Stage 0 (2026-05-12 → 2026-05-16). STAGE_1_PLAN.md is the parallel live execution plan for Stage 1. The two coexist: STAGE_0_PLAN.md is now a historical artifact pointing forward (see its closing note); STAGE_1_PLAN.md is the working surface. Each subsequent stage gets its own plan doc.

This doc mirrors STAGE_0_PLAN.md structurally — Pre-flight items, success contract, failure handling, weekly sequence, maker/checker assignments, compounding, success artifacts, escalation. Differences from STAGE_0_PLAN.md reflect Stage 1's customer-facing reality (commercial gate alongside technical gate; more components; longer phase) and the Path 2 discipline calibration per ADR-0006.

## 1. Pre-flight items completed before Stage 1 week 1

- ✅ Stage 0 closeout draft open (PR #21) — converts to ready-for-review after senior-engineer walkthrough.
- ✅ Stage 0 hygiene parallel-track scheduled — first-customer run, cold-engineer protocol doc, technical reference doc, video walkthrough recording.
- ✅ Senior-engineer walkthrough session being scheduled per CLAUDE.md §Stage 0 process notes trigger.
- ✅ PR-E (this PR) ships ADR-0006 + STAGE_1_PLAN.md + Drift-15/-16/-17/-18 plan-doc resolutions + Brain substrate code [A].

## 2. Stage 1 success contract (what must be true on day 70, ~2026-07-25)

1. **5 paying tenants** at $99–$1,499/month each (Tier 1 / 2 / 3 per canonical plan §6) sustaining the gate signal:
   - At least one marketing artifact shipped per day for 2 consecutive weeks.
   - Documented ≥50% reduction in their existing tooling spend (Buffer / Hootsuite / Jasper / Copy.ai / equivalent).
2. **§3.5 perf gate passed per tenant.** Five dimensions per canonical plan §3.5:
   - Wall-clock <6 hours for the tenant's onboarding pass.
   - LLM spend <$50 per onboarding.
   - Org-snapshot accuracy >90% of operator-reviewed claims.
   - Skill recommendation precision >80% of recommendations kept after operator review.
   - Destructive-action incidents == 0 across all onboardings.
3. **Full Onboarding Crew running autonomously** on every new tenant (six agents from canonical plan §3.2).
4. **Web admin live** with org-chart UI, ticket queue, calendar, budget dashboard, Brain explorer.
5. **Five marketing agents live** (CMO + Content Writer + Social Manager + Ad Ops + Growth Analyst).
6. **~40 marketing Skills mirrored** from gooseworks + CosmoBlk + custom Galileo Skills.
7. **React Native mobile app + Telegram bot** approval surfaces live.
8. **Promptfoo eval on every Skill PR.**
9. **Destructive-action lockdown enforced platform-wide** via Temporal signal gate + Mirage snapshot-before-write.
10. **Connectors at 8** (GDrive, Slack, Gmail, X, LinkedIn, GA, Plausible, Stripe; current Stage 0 set is 3: github, slack, gdrive).

If any of items 1–10 is not satisfied on day 70, [`docs/closeouts/CLOSEOUT_STAGE1.md`](../closeouts/CLOSEOUT_STAGE1.md) is written naming the structural finding and the adjustment-ladder cut taken.

## 3. The 14 components and their prerequisite chain

Per ADR-0006 and the Stage 1 structural reading round. The dependency chain determines order; items in the same row can ship in parallel.

```
Stage 0 main (HEAD = 8db8158 post-PR #20)
            │
            ▼
   [A] Brain substrate (pgvector + AGE on Postgres)        ◄── PR-E ships this
            │
            ├──────────────┬──────────────┬──────────────┐
            ▼              ▼              ▼              ▼
        [B] Ingestion   [C] Org-Mapper [D] Skill        (parallel-track:
            Agent            Agent         registry +        [J] connector
                                           Skill-            expansion 3→8;
                                           Selector          [N] Promptfoo
                                                             eval; [M]
                                                             Supabase Auth)
            │              │              │
            ▼              ▼              ▼
                  [E] QA Agent
                          │
                          ▼
            [F] Operator-review UI ──── [G] Approval-state machine
                          │              (Temporal signal gate)
                          │              │
                          ▼              ▼
                Onboarding Crew GA — six agents satisfying §3.5 perf gate
                          │
                          ├──────────────┬──────────────┐
                          ▼              ▼              ▼
                   [H] Five           [I] Skill      [K] Mobile +
                       marketing          format +       Telegram
                       agents             ~40 Skills     approval
                          │
                          ▼
        [L] Destructive-action lockdown enforcement
            (uses [G] signal gate + Mirage snapshot-before-write)
                          │
                          ▼
            Stage 1 gate: 5 paying tenants + §3.5 per tenant
```

## 4. Per-component PR projection with Path 2 calibration

Path 2 column is explicit (per ADR-0006 + Pushback Item 5 of the PR-E verification round): **Full** = full read-first + verification + inline plan + diff readback discipline; **Streamlined** = pattern reuses an established shape with shorter planning rounds.

| # | Component | PRs | Path 2 | Notes |
|---|---|---|---|---|
| [A] | Brain substrate (pgvector + AGE) | 1 (this PR-E) | **Full** | First custom Dockerfile; first cross-extension Postgres image |
| [B] | Ingestion Agent | 3 | **Full** (per format family) | Docling for office docs; MarkItDown for everything else; Whisper for audio |
| [C] | Org-Mapper Agent | 2 | **Full** | Synthesis workflow + JSON schema/prompts |
| [D] | Skill registry + Skill-Selector | 3 | **Full** | Go service + Python agent + calibration artifact for §3.5 #4 |
| [E] | QA Agent | 2 | **Full** | Cross-check workflow + confidence-scoring/flags |
| [F] | Operator-review UI | 4 | **Full** (UI architecture) | Org-chart; ticket queue; Brain explorer; budget dashboard |
| [G] | Approval-state machine | 1 | **Full** | Temporal signal gate + web→agent-runner contract |
| [H] | Five marketing agents (CMO + 4 specialists) | 5 | **Full** for CMO; **Streamlined** for specialists 2-5 | CMO sets the pattern; specialists reuse |
| [I] | Skill format + ~40 Skills | 3 | **Full** for format adoption; **Streamlined** for Skill batches | Mirror gooseworks + CosmoBlk + custom |
| [J] | Connector expansion (3→8) | 5 | **Streamlined** | Each reuses PR-D dispatch pattern; Gmail / X / LinkedIn / GA / Plausible / Stripe |
| [K] | Mobile + Telegram approval | 3 | **Full** for first PR; **Streamlined** for subsequent | Expo skeleton + approval inbox + Telegram bot |
| [L] | Destructive-action lockdown | 2 | **Full** | First Mirage Layer-5 adopter; signal-gate enforcement |
| [M] | Supabase Auth | 2 | **Full** | Replaces dev JWT keypair; substantial walkthrough rewrite. **Verify Supabase Auth schema interactions with ag_catalog-prefixed search_path during component planning** (PR-E `ALTER DATABASE` from [A] sets the default search path). |
| [N] | Promptfoo eval harness | 1 | **Streamlined** | Single CI integration; mostly mechanical |
| **Subtotal** | | **37 PRs** | | |
| Plan-deviations expected (~1 per architectural week × 6 weeks) | | **+6 PRs** | **Full** | Each is a four-part deviation in one indivisible PR |
| Stage 0 hygiene parallel-track | | **+3-5 PRs** | **Streamlined** | Walkthrough refinements, cold-engineer protocol, technical reference doc |
| Drift-resolution and small fixes | | **+3-5 PRs** | **Streamlined** | Continued Drift-19+ resolutions, CI tweaks |
| **Total estimate** | | **~48-50 PRs** | | ~8 PRs/week if compressed evenly; realistic distribution is heavier in weeks 1-3 (foundational components) and weeks 5-6 (marketing agent buildout) |

## 5. Six-week calendar — weekly milestones

| Week | Window | Milestone | Components landing |
|---|---|---|---|
| **Week 1** | 2026-05-16 → 2026-05-23 | Brain substrate live; Ingestion Agent shape locked | [A] this PR; [B] PRs 1-2 (Docling + MarkItDown); [J] one connector if parallel-track bandwidth allows |
| **Week 2** | 2026-05-23 → 2026-05-30 | Ingestion complete; Org-Mapper shape locked | [B] PR 3 (Whisper); [C] PRs 1-2; [J] one more connector |
| **Week 3** | 2026-05-30 → 2026-06-06 | Onboarding Crew six agents online; first review trigger | [D] PRs 1-3; [E] PRs 1-2; **Week-3 review checkpoint** — ladder rung 1 if behind |
| **Week 4** | 2026-06-06 → 2026-06-13 | Operator-review UI shipping; approval gate live | [F] PRs 1-4; [G]; **Week-4 review checkpoint** — ladder rung 2 if behind |
| **Week 5** | 2026-06-13 → 2026-06-20 | Marketing agents online; Skills mirrored | [H] PRs 1-3 (CMO + 2 specialists); [I] PRs 1-2; **Week-5 review checkpoint** — ladder rung 3 if behind |
| **Week 6** | 2026-06-20 → 2026-06-27 | Approval surfaces + lockdown enforcement + auth | [H] PRs 4-5; [I] PR 3; [K] PRs 1-3; [L] PRs 1-2; [M] PRs 1-2; [N] |
| **Weeks 7-10** | 2026-06-27 → 2026-07-25 | Commercial gate satisfaction; 5 paying tenants × 2 weeks daily-artifact signal | No new architectural components — operational only |

The Week-N review checkpoints are when the adjustment ladder fires per ADR-0006 if velocity demands.

## 6. Adjustment ladder (cross-reference to ADR-0006)

Named order-of-cuts if velocity demands intervention. Triggers on velocity signals at week-N review checkpoints (not on a single bad week):

1. **Week 3.** Drop React Native mobile [K] (Telegram covers approval surface). Stage 2 carries it.
2. **Week 4.** Drop marketing agents 4-5 [H4-H5] (CMO + 3 specialists covers the gate's daily-artifact signal). Stage 2 carries Ad Ops + Growth Analyst.
3. **Week 5.** Reduce Skills count from ~40 to ~25 (highest-value-first: gooseworks GTM core + CosmoBlk email-marketing-bible essentials).
4. **Last resort.** Defer Promptfoo eval [N] entirely to Stage 2; Skill versioning still occurs without eval-on-every-PR.

## 7. Pre-registered failure handling

Same shape as Stage 0 PRP rule 1. Each component's failure path is the closeout doc that names the structural finding.

| Failure | Closeout doc | Probable adjustment-ladder rung |
|---|---|---|
| Component [A] Brain substrate doesn't build on the chosen Postgres image | `CLOSEOUT_BRAIN_SUBSTRATE_RECONSIDERED.md` | None — would be a plan-deviation in PR-E itself; addressed at planning time per ADR-0006 verification round outcomes |
| [B] Ingestion latency exceeds §3.5 wall-clock cap | `CLOSEOUT_INGESTION_LATENCY.md` | Reduce per-source doc cap; or upgrade VM tier in Stage 2 |
| [C] Org-Mapper accuracy <90% | `CLOSEOUT_ORG_MAPPER_PRECISION.md` | Org-Mapper v2 spec; not a Stage 1 ladder cut |
| [D] Skill-Selector precision <80% | `CLOSEOUT_SKILL_SELECTOR_CALIBRATION.md` | Re-calibrate before redeploy; calibration artifact committed |
| Velocity < target at week 3/4/5 review | n/a (use the adjustment ladder) | Ladder rung 1 / 2 / 3 per above |
| Commercial gate misses (fewer than 5 paying tenants by day 70) | `CLOSEOUT_STAGE1_GATE.md` | Phase extends by 4 weeks; structural finding documented |
| Any destructive-action incident | Platform-wide pause + per canonical plan §3.5 row 5 | Not negotiable; ladder doesn't apply |

## 8. Stage 1 deviations — expected cadence ~6 across the phase

Per Stage 0 closeout pattern #1: the project ships at a cadence of ~1 plan-deviation per architectural week. Stage 1 has 6 architectural weeks, so expect ~6 plan-deviations. **Zero deviations across multiple weeks indicates discipline skipping**, not perfect planning — per Stage 0 closeout §7 framing.

Each deviation gets the four-part template (closeout + ADR + plan edits + code in one indivisible PR). Drift-numbering continues from Drift-19 (Drift-15/-16/-17/-18 land in PR-E itself).

## 9. Compounding (v7 rule 7)

Stage 0 produced five solutions docs. Stage 1 will compound these and add new ones. Expected Stage 1 solutions-doc additions (named in advance so we look for the pattern rather than miss it):

- **`SOLUTION_BRAIN_INGESTION_QUALITY.md`** — patterns for getting clean markdown out of heterogeneous sources; latency/cost trade-offs. First instance lands during [B].
- **`SOLUTION_AGENT_QUALITY_VS_LLM_COST.md`** — first encounter with the LLM-spend cap as a real constraint vs. Stage 0's gate test budget. First instance during [H] CMO agent.
- **`SOLUTION_MULTI_CHANNEL_APPROVAL.md`** — web admin + mobile + Telegram all wire to the same Temporal signal gate. First instance during [K].
- **`SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md`** — fifth and subsequent instances of the existing pattern, likely along axes already named (deployment topology / container topology / ecosystem utilities / distribution channel) or new ones.

The discipline: every non-obvious finding during Stage 1 becomes a solutions doc within 72 hours of surfacing. Cumulative count is part of Stage 1's closeout artifacts.

## 10. Maker / checker assignments (v7 rule 5)

Same shape as STAGE_0_PLAN.md §5. Stage 1 makes the maker/checker symmetry explicitly *across PRs* in a phase rather than per-PR-in-isolation:

- **Maker** is the agent (Claude Opus 4.7) authoring code or docs.
- **Checker** is Emmanuel reviewing the PR. For UI work, `/qa` runs against a real browser by a separate Claude session.
- **Inline planning is part of the maker/checker discipline.** Every Full-discipline PR gets an inline plan + diff readback round. Streamlined PRs get a lighter round (file list + tl;dr) but never skip the maker/checker separation.
- **Calibration artifacts (Skill-Selector precision, Org-Mapper accuracy) need explicit checker sign-off** before the agent is redeployed per §7 above.

## 11. Risks

Two big ones:

1. **Pace sustainability.** 6 weeks at ~8 PRs/week is a pace increase over Stage 0 (Stage 0 averaged ~5 PRs/week across 4 weeks). Stage 1's PRs are also individually larger on average (the marketing agents alone are 5 PRs of non-trivial complexity). The adjustment ladder is the safety net; the discipline is *checking velocity at week-N reviews, not after the phase ends*.

2. **Commercial gate calendar is uncompressible.** 5 paying tenants × 2 weeks daily-artifact + 50%-tooling-spend-reduction needs calendar time after the technical surface ships. If technical surface slips, gate satisfaction slips proportionally. STAGE_1_PLAN.md encodes the honest 10-week target; mid-phase slippage extends the day-70 milestone rather than compressing the gate.

## 12. Sign-off

This plan is the working surface for Stage 1. Edits to this plan are PRs and require checker approval. Plan-doc edits as part of Drift-N resolutions (continued from Drift-19) land in the PR that surfaces them, same discipline as Stage 0.

Stage 1 starts when PR-E merges and Stage 0 closeout PR #21 converts from draft to ready-for-review (post senior-engineer walkthrough).
