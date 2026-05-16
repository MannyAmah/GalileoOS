# CLOSEOUT — Stage 0 (2026-05-12 → 2026-05-16 code/docs; gate-test pending)

**Phase:** Stage 0 — kernel + agent-runner + observability + Onboarding Crew scaffold.
**Outcome:** Code and docs shipped through PR #20 (Week 4 PR-D). Gate-test results pending senior-engineer install-walkthrough session.
**Date drafted:** 2026-05-16 (post-PR-D merge).
**Status:** **Two-phase closeout.** Sections 1–4, 6, 7 finalize at draft time. Section 5 holds the senior-engineer walkthrough results and fills in once the session lands; this doc converts from draft to ready-for-review at that point.

This is structurally different from the prior closeouts. [`CLOSEOUT_LAYER3_MIRAGE_RECONSIDERED.md`](CLOSEOUT_LAYER3_MIRAGE_RECONSIDERED.md), [`CLOSEOUT_OPIK_RECONSIDERED.md`](CLOSEOUT_OPIK_RECONSIDERED.md), and [`CLOSEOUT_MCP_PER_SOURCE_RECONSIDERED.md`](CLOSEOUT_MCP_PER_SOURCE_RECONSIDERED.md) are *plan-deviation closeouts* — one finding, one decision, one PR. This closeout is a *phase closeout* — everything that happened across four weeks. Different shape, different length.

## 1. What Stage 0 shipped

**19 PRs merged through main**, organized by week:

| Week | PRs | What landed |
|---|---|---|
| Week 0 (bootstrap) | #1, #3, #4, #5, #6, #7 | Repo bootstrap, plan + STAGE_0_PLAN draft, CLAUDE.md + AGENTS.md operating discipline, ADR-0001 (repo namespace) + ADR-0002 (protection approval relaxation), monorepo skeletons, minimal CI, buf breaking ref fix |
| Week 1 (CI + discipline) | #8, #9, #10, #11, #12 | CI expansion (Go + Python + Web + protobuf), devcontainer + Makefile, latest-1 Go posture pin, Mirage probe apparatus (calibration artifact / 1st size exception), `SOLUTION_CI_GUARD_VERIFICATION.md` + `SOLUTION_DEPENDENCY_COUPLING.md` |
| Week 2 (1st plan-deviation) | #13, #14 | Mirage Layer 3 → Layer 5 relocation (1st plan-deviation, ADR-0003, closeout); `kernel/probe/mirage` → `kernel/probe/connector` rename |
| Week 3 (kernel + observability + agent-runner) | #15, #16, #17, #18, #19 | PR-A gateway compose stack + JWT-verified LiteLLM passthrough (runtime-introduction / 2nd size exception); `SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md` (3 instances); PR-B cost-meter + budget cap + `cost_events` ingest + Jaeger/OTel substitution (2nd plan-deviation, ADR-0004, plan-deviation-with-code / 3rd size exception); PR-C Hello Agent end-to-end; PR-C follow-up workflow unit tests |
| Week 4 (Onboarding Crew scaffold) | #20 | PR-D per-source MCP dispatch + Connector + Crawler + Temporal worker + credentials store + manifest-check Go binary + install walkthrough doc (3rd plan-deviation, ADR-0005, plan-deviation-with-code 2nd instance / 4th size exception) |

**Surface area at the close of code/docs work:**

- **Kernel (Go):** gateway with JWT verification, budget-cap middleware, `cost_events` ingestion + Stripe reconciliation, OTel tracer + Jaeger backend, inline migration runner, embedded SQL migrations; agent-runner with Temporal workflow orchestration; manifest-check binary validating five §3.5 gate dimensions; jwt-tool dev keypair generator.
- **Agents (Python):** Onboarding Crew scaffold — Connector + Crawler Temporal activities, `ConnectorWorkflow` + `CrawlerWorkflow` on the `galileo-onboarding-crew` task queue, encrypted credentials store (HKDF-SHA256 → AES-256-GCM with AAD binding), CLI with YAML config. Hello-Agent worker on the `galileo-agent-runner` task queue.
- **Web (Next.js):** Hello Agent end-to-end UI submitting goals to the agent-runner, polling task results, surfacing `cost_events.request_id` correlation.
- **Schemas:** protobuf v1 contracts under `schemas/galileo/v1/*.proto`, generated Go code committed under `kernel/gen/galileo/v1/`.
- **Deploy:** `docker-compose.yml` with Postgres 17.9 + Temporal 1.29.6.1 + LiteLLM v1.83.14-stable.patch.3 + Jaeger 2.18.0 + OTel collector 0.152.0; `github-mcp-server:v1.0.4` as a per-invocation Docker subprocess.
- **Docs:** authoritative plan + STAGE_0_PLAN; five ADRs (0001–0005); three plan-deviation closeouts; five solutions docs; one install walkthrough; CLAUDE.md + AGENTS.md operating contracts.

**Artifact counts (across docs/):**

| Artifact type | Count | Files |
|---|---|---|
| ADRs | 5 | `docs/decisions/0001` through `0005` |
| Plan-deviation closeouts | 3 | Mirage, Opik, MCP-per-source |
| Solutions docs | 5 | CI guard verification, dependency coupling, CI expansion findings, dependency shape verification (4 instances), coverage drill-down |
| Stage plans | 1 | `STAGE_0_PLAN.md` (authoritative live doc) |

## 2. The three plan-deviations and their resolutions

Each followed the four-part template (closeout + plan edits + ADR + code in one indivisible PR) and preserved the underlying intent while changing the implementation.

### PR #13 — Mirage Layer 3 → Layer 5 (2026-05-13, ADR-0003)

**Surface:** The plan placed Mirage at Layer 3 (Go kernel) as the unified data-plane substrate. Discovery: Mirage ships Python and TypeScript SDKs only, with no Go SDK and no native server. Layer 3 placement would have required a permanent Python sidecar not named in the plan.

**Resolution:** Mirage relocated to Layer 5 (integrations), agent-side. The live probe became unnecessary — the structural mismatch was already decisive. Stage 1 trigger: first agent-side adopter that imports `mirage-ai` in-process.

**Instance 1 of the dependency-shape pattern** (axis: deployment topology).

### PR #17 — Opik → Jaeger + OTel (2026-05-14, ADR-0004)

**Surface:** The plan named Opik as the observability substrate at Layer 5, implicit framing of a single observability service. Discovery: Opik's self-hosted deployment is a 13-container platform (MySQL + Redis + ClickHouse + Zookeeper + MinIO + 5 Opik-specific services + Jaeger + OTel collector).

**Resolution:** Substituted Jaeger 2.18.0 + OTel collector 0.152.0 — two containers covering the trace ingest + storage + query surface the gate test requires. Gate-test contract preserved unchanged (100 demo runs → 100 parent spans in the backend). Stage 1 trigger: customer requires Opik's LLM-evaluation surface natively; or scale exceeds Jaeger's pluggable-storage limits.

**Instance 2 of the dependency-shape pattern** (axis: container topology). **First size exception in the plan-deviation-with-code category.**

### PR #20 — Per-source MCP dispatch (2026-05-16, ADR-0005)

**Surface:** The plan named three reference MCP servers (`@modelcontextprotocol/server-{github,slack,gdrive}`) invoked via `npx -y`. Discovery: all three reference packages were archived upstream on commit `d53d6cc75c` on 2025-05-29; GitHub's vendor-maintained replacement (`github/github-mcp-server` v1.0.4) is Docker/Go-binary distributed, not npm.

**Resolution:** Per-source dispatch by `source_kind`. Docker subprocess MCP (with `--init`) for github; `slack_sdk` and `google-api-python-client` direct for slack and gdrive. Four reversal triggers in ADR-0005 (Slack vendor MCP, Google vendor MCP, sixth source-kind, Docker unavailability).

**Instance 4 of the dependency-shape pattern** (axis: distribution channel — newly named). **Second size exception in the plan-deviation-with-code category.**

### The pattern that emerged from the three together

Read dependency → find mismatch → substitute structurally → name reversal triggers → encode in ADR + closeout + plan edits + code, all in one PR. The pattern is now durable enough that PR-D's discovery pass surfaced the fourth instance *before* implementation rather than after, validating the upstream discipline (the `SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md` doc) as load-bearing rather than retrospective.

## 3. Discipline patterns that emerged

Worth being explicit so Stage 1 inherits these rather than rediscovering them. Nine patterns documented to date:

1. **Dependency-shape verification.** `SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md` with the four-axis checklist (deployment topology / container topology / ecosystem utilities / distribution channel) and one verification command per axis. Compounded across four instances (Mirage, Opik, wolfi-base, github-mcp-server). PR #16 turned it into a doc; PR-D turned it into a structured checklist.

2. **Maker/checker symmetry on estimate quality.** First-instance complex code consistently undershoots my (the maker's) line projections. The honest variance disclosure surfaced in PR-A (`migrate.go`: 160 vs ~30 projected), PR-C (`hello-agent.tsx`), and PR-D (`crawler.py` 264 vs 180, `manifest-check/main.go` 213 vs 120, install walkthrough 161 vs 120). **Stage 1 estimate adjustment: apply a 1.3–1.5x multiplier on first-instance complex-component projections.**

3. **Read-first discipline.** Three reading passes precede every non-trivial PR — structural reading round before the inline plan; inline plan iteration with reviewer before drafting code; diff readback before opening the PR. Plus one reading pass on the merge side: line-level review against the approved shape, not against the diff in isolation. The discipline collapses review surface from "did this work" to "did this match the agreed plan."

4. **Drift-numbering sequence.** Each planning miss gets a sequential number; resolutions encoded as plan-doc edits in the PR that surfaces them. Stage 0 surfaced Drift-1 through Drift-14 across four weeks; their resolutions are scattered across PR-A through PR-D's commits. Cumulative drift count = the project's planning-error ledger. Numbering reset between stages is a Stage 1 decision worth taking deliberately.

5. **Size-exception discipline.** Qualitative, not quantitative. Named categories cite prior precedents; new categories require deliberate naming + structural argument. Three categories observed in Stage 0:
   - **Calibration artifact** (PR #10): apparatus + mocks + self-validation tests indivisible.
   - **Runtime introduction** (PR #15 / PR-A): code + integration test + compose stack indivisible.
   - **Plan-deviation with code** (PR #17 / PR-B, PR #20 / PR-D — now a recurring shape with two instances): closeout + plan edits + ADR + deviation code indivisible.

6. **Co-change ecosystem discipline.** Related packages within an upstream ecosystem move together in one PR. Examples: Temporal SDK + API (`go.temporal.io/sdk` + `go.temporal.io/api`); OTel SDK + exporters (`opentelemetry-sdk` + `opentelemetry-exporter-otlp`); CI pin + devcontainer pin. Drift between ecosystem siblings produces silent breakage that the test surface doesn't always catch.

7. **Reconsideration triggers.** Every Stage 0 simplification names a structural trigger for Stage 1 reconsideration rather than a time-based deadline. Surfaced in §4 below as the deferred-work ledger.

8. **Honest variance disclosure.** Every size overshoot decomposes the variance into estimate-realism vs scope drift; the checker (Emmanuel) flags maker-estimate quality when wrong rather than absorbing the variance silently. Establishes the explicit maker/checker symmetry that pattern #2 names.

9. **New-test-category needs approval.** PR-C's `workflow_test.go` trim and PR #19 follow-up established this: when implementation surfaces a new test category not in the inline plan, the test gets explicit approval rather than absorbed silently. Future test categories appearing mid-implementation get surfaced as a fork, not added. *Pattern status: one instance to date (PR-C → PR #19). Listed as a numbered pattern because the structural argument is sound, but it earns full pattern status when a second instance lands.* This honest framing is itself discipline-strengthening — eight well-established patterns and one emerging one is more accurate than pretending all nine have equal weight.

## 4. What's deferred to Stage 1 with concrete triggers

Every Stage 0 simplification surfaces here with the trigger that fires reconsideration. **No time-based deferrals — every entry names a structural condition.**

| Deferral | Trigger | Origin |
|---|---|---|
| pgvector + AGE wiring for the Brain | Stage 1 Brain scope opens. First Brain consumers: Ingestion Agent + QA Agent (no separate threshold trigger), and Org-Mapper Agent (with the additional threshold trigger in row 3 below). | ADR-0003 scope note; plan §3.3 lists Ingestion + Org-Mapper + QA as Brain-dependent agents |
| Skill registry + Skill-Selector Agent | Stage 1 §3.5 threshold #4 (Skill recommendation precision >80%) becomes evaluable | `manifest-check` `[gate]` Skill-precision line; plan §3.5 |
| Org-Mapper Agent + org-snapshot accuracy threshold | Stage 1 §3.5 threshold #3 becomes evaluable beyond Stage 0's per-source coverage proxy | plan §3.5; `manifest-check` org-snapshot dimension uses source-kind coverage as the Stage 0 proxy |
| Operator-review UI replacing CLI invocation | Stage 1 spec §6 deliverable; the Onboarding Crew CLI is the Stage 0 invocation surface only | ADR-0005 + plan Week 4 deliverable revision |
| Real-GitHub CI fixture with GitHub Actions secrets management | (a) Stage 1's first multi-agent github usage that modifies connector dispatch, OR (b) first observed regression in the github dispatch path that the walkthrough catches | CLAUDE.md §Stage 0 process notes (PR-D context) |
| Mirage Layer 5 first agent-side adopter | First agent-side import of `mirage-ai` for in-process source mounting | ADR-0003 forward path |
| Full observability platform reconsideration | (a) customer requires Opik's LLM-evaluation surface natively; (b) trace volume exceeds Jaeger's pluggable-storage limits | ADR-0004 reversal triggers |
| 402 → 429 reconsideration | First external API consumer expects HTTP 429 for budget-cap responses | `kernel/cmd/gateway/budget.go` reconsideration note |
| snake_case → camelCase reconsideration | First external API consumer expects camelCase JSON convention | `web/lib/api-types.ts` reconsideration note |
| Hand-written TS types → generated TS (bufbuild/es) | (a) web complexity grows to 3+ entity types, OR (b) a second TypeScript client appears | `web/lib/api-types.ts` |
| Inline migration runner → golang-migrate | (a) first down-migration need, (b) first multi-environment migration drift, OR (c) first PITR-aware restore | CLAUDE.md "Migration tooling" triggers |
| Postgres `tenant_credentials` → Supabase Vault or Infisical | First production deployment lands and the dev Ed25519 keypair is no longer the right key-management handle | ADR-0005 implicit; Drift-8 resolution |
| Agent-side snapshot mechanism → Temporal-signal gate enforcement | First Stage 1 agent path that opens a destructive write surface | CLAUDE.md §Destructive-action lockdown; spec §3.3 step 5 |

13 entries. Each one is a structural trigger, not a deadline. Stage 1 starts with this ledger as the prioritization input.

## 5. Stage 0 gate-test results (reserved)

> **Status:** Pending senior-engineer walkthrough session. This section converts the closeout from draft to ready-for-review when the session lands.

**Structure to populate (per `STAGE_0_PLAN.md` §Week 4 exit criterion):**

- Three teammate runs (Emmanuel + two volunteers), each from a fresh Ubuntu 24.04 VM (or equivalent), names and dates.
- Time-to-complete per teammate (target ≤ 30 minutes per `docs/onboarding/install_walkthrough.md`).
- Hello Agent verification: 100 runs, 100 traces in Jaeger, cost matches Stripe to the cent.
- Onboarding Crew verification: `manifest-check` exits 0 against the internal test workspace.
- Deviations from the walkthrough doc, with proposed doc fixes if any.
- Senior-engineer's findings (identification artifact: name + written commitment + cold-state per CLAUDE.md §Stage 0 process notes).

**Trigger to fill this section:** the cold-engineer session runs to completion. The walkthrough video URL gets backfilled into `docs/onboarding/install_walkthrough.md` at the same time.

## 6. Honest accounting of what almost went wrong

Worth including in a phase closeout that doesn't appear in per-PR closeouts. Two near-misses, both caught at planning time, both would have been multi-week rework if caught at implementation time:

### Near-miss 1 — Mirage Layer 3 placement

**What almost shipped:** The original plan placed Mirage as the Go kernel's data-plane substrate at Layer 3. Implementation under that placement would have required either (a) a Go-language SDK for Mirage (does not exist), (b) a permanent Python sidecar bridging the kernel to Mirage (not in the plan, multi-week scope), or (c) static-linking a Python interpreter into a Go binary (not viable).

**What caught it:** The structural reading round before the inline plan's Week 1 deliverables. Reading Mirage's docs surfaced "ships Python and TypeScript SDKs only, with no Go SDK and no native server." The structural mismatch was decisive — a live probe couldn't change the language coverage.

**Estimated rework if caught at implementation time:** ~3 weeks. A Python sidecar deployment plus the contract surface across Layer 3 ↔ Layer 5 would have been net-new architecture, not a refactor.

### Near-miss 2 — Opik 13-container assumption

**What almost shipped:** PR-A's compose stack named Opik as the Layer 5 observability deliverable, framed implicitly as a single service container comparable to Postgres or LiteLLM. Implementation under that framing would have produced a `deploy/compose/docker-compose.yml` adding five state stores (MySQL + Redis + ClickHouse + Zookeeper + MinIO) and eight services — a stack roughly twice the size of the entire kernel.

**What caught it:** The Week 3 inline-plan discovery pass for PR-A. Reading Opik's self-hosted deployment docs surfaced the 13-container count. The mismatch was structural — no traffic volume would change the container topology of a supported deployment.

**Estimated rework if caught at implementation time:** ~2 weeks. The compose-stack PR would have merged with five state stores, the Stage 0 gate test would have run against a state-store sprawl that no Stage 0 feature actually needs, and the substitution would have happened at Stage 1 against operational debt rather than against a clean slate.

### What prevented both

The pattern that prevented these is the dependency-shape verification doc, which **didn't exist as a doc until PR #16 surfaced the third instance** (wolfi-base utility absence). The first two near-misses *produced* the discipline; the third near-miss (caught at CI time, not planning time) *codified* the discipline; the fourth near-miss (github-mcp-server distribution channel, PR-D) *applied* the discipline at planning time as designed.

Stage 0's biggest invariant carry-over to Stage 1 is therefore not a code surface — it is the read-first discipline that produced the dependency-shape verification doc. Stage 1's first dependency adoption should reach for the four-axis checklist before naming the dependency in the plan, not after.

## 7. What comes next

Stage 1 starts with §4's deferred-work ledger as the prioritization input. Two dimensions of priority:

**Architectural consequence (biggest first):**

1. Skill registry + Skill-Selector Agent — unlocks plan §3.5 threshold #4 (Skill precision >80%), which `manifest-check` currently emits as a `[gate] N/A` line per ADR-0003.
2. Org-Mapper Agent + the org-snapshot accuracy threshold — unlocks plan §3.5 threshold #3 (org-snapshot coverage >90%); the Stage 0 proxy (source-kind coverage) is structurally distinct from real org accuracy.
3. pgvector + AGE Brain wiring — substrate for both #1 and #2.
4. Operator-review UI — replaces CLI invocation surface and unlocks Stage 1 destructive-action gating flows (per CLAUDE.md §Destructive-action lockdown).

**Tactical (smallest first, prevents future regression):**

1. Real-GitHub CI fixture with `GITHUB_FIXTURE_PAT` secrets-management — small PR, prevents github dispatch-path regression.
2. Inline migration runner → golang-migrate — only if any of the three triggers fires.
3. Hand-written TS types → generated TS — only if the web entity-count or client-count triggers fire.

The third plan-deviation closing in Stage 0 also clarifies a meta-pattern: **the project ships at a cadence where one major plan-deviation lands per week.** Mirage in Week 2, Opik in Week 3, MCP per-source in Week 4. Each surfaced via the structural-reading discipline before code drafting; each preserved the underlying gate-test contract while changing the implementation. If Stage 1 sustains this cadence, the dependency-shape verification doc will compound to a fifth instance within the first Stage 1 inline plan — and that's the load-bearing signal that the discipline is working, not failing.

**Stage 1 should expect ~1 plan-deviation per architectural week.** Zero across four weeks indicates the discipline is being skipped, not that the planning is suddenly perfect. The framing reversal is the substantive one: deviations count as **findings the discipline produced against real dependencies**, not as rework incurred from bad planning. Future Stage 1 planning conversations should reach for this framing when a fourth and fifth plan-deviation surface, not panic about them.

## Cross-references

- **Plan:** [`docs/plans/STAGE_0_PLAN.md`](../plans/STAGE_0_PLAN.md) — authoritative Stage 0 execution plan, updated through PR-D.
- **ADRs:** [`docs/decisions/0001`](../decisions/0001-repo-namespace.md) through [`0005`](../decisions/0005-mcp-per-source-vs-mixed.md).
- **Plan-deviation closeouts:** [`CLOSEOUT_LAYER3_MIRAGE_RECONSIDERED.md`](CLOSEOUT_LAYER3_MIRAGE_RECONSIDERED.md), [`CLOSEOUT_OPIK_RECONSIDERED.md`](CLOSEOUT_OPIK_RECONSIDERED.md), [`CLOSEOUT_MCP_PER_SOURCE_RECONSIDERED.md`](CLOSEOUT_MCP_PER_SOURCE_RECONSIDERED.md).
- **Solutions docs:** [`SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md`](../solutions/SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md) (the load-bearing one), [`SOLUTION_CI_GUARD_VERIFICATION.md`](../solutions/SOLUTION_CI_GUARD_VERIFICATION.md), [`SOLUTION_DEPENDENCY_COUPLING.md`](../solutions/SOLUTION_DEPENDENCY_COUPLING.md), [`SOLUTION_CI_EXPANSION_FINDINGS.md`](../solutions/SOLUTION_CI_EXPANSION_FINDINGS.md), [`SOLUTION_COVERAGE_DRILL_DOWN.md`](../solutions/SOLUTION_COVERAGE_DRILL_DOWN.md).
- **Walkthrough doc:** [`docs/onboarding/install_walkthrough.md`](../onboarding/install_walkthrough.md) — the cold-engineer flow that §5 will exercise.
