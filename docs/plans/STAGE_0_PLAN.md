# Stage 0 Plan — Galileo OS

| Field | Value |
| --- | --- |
| **Stage** | Stage 0 — Foundations |
| **Window** | Weeks 1–4 (kickoff: 2026-05-12) |
| **Audience** | Internal only. No paying customers. |
| **Headline deliverable** | "Hello Agent" running durably with traces and budget, Mirage layer-relocation closeout committed (Reading 2 — Layer 5 agent-side; see PR #13 and ADR-0003), Stage 0 gate passed. |
| **Authoritative spec** | [`docs/galileo_os_infrastructure_plan.md`](../galileo_os_infrastructure_plan.md) §6 Stage 0 + Appendix B + §3.5 + §2.4 |
| **Author** | Claude Opus 4.7 (1M ctx), under Emmanuel's direction, 2026-05-12 |
| **Status** | DRAFT — awaiting sign-off before Week 1 begins |

---

## 0. Why this document exists

The project kickoff requires this plan to be drafted, reviewed, and signed off **before** any Week 1 implementation keystrokes (v7 rule 6: plan before code). The plan in `docs/galileo_os_infrastructure_plan.md` is the authoritative spec; this document is the operational sequence that executes Stage 0 against that spec without softening it.

If Stage 0 succeeds, the result is four artifacts (§7 below). If any of them fail, this plan does **not** soften the gate — it produces a closeout naming the structural finding (v7 rule 9) and Stage 1 does not begin.

## 1. Pre-flight items completed before Week 1

- [x] **Plan read in full.** All 1185 lines of `docs/galileo_os_infrastructure_plan.md`. Eight-layer architecture, Day-Zero Onboarding Crew, v7 discipline, Stage 0 gate (§6), and Appendix B docker-compose understood.
- [x] **Global `~/.claude/CLAUDE.md` honored.** Galileo is treated as a standalone product per kickoff directive; no cross-pollination with Livemore / Alpha Sentinel / Xpedinet etc.
- [x] **Repo confirmed: `github.com/MannyAmah/GalileoOS`.** Apache 2.0. Plan deviation from `galileoos/galileo-os` accepted by Emmanuel on 2026-05-12; logged in [`docs/decisions/0001-repo-namespace.md`](../decisions/0001-repo-namespace.md) (to be authored Week 1).
- [x] **Local working tree initialized.** `/Users/emman/GalileoOS` is a git repo tracking `MannyAmah/GalileoOS` `main`. LICENSE retained. The plan moved into `docs/`. `.gitignore` in place.
- [x] **Memory bootstrapped.** Project facts, v7 discipline, locked decisions, Stage 0 sequence, and the standalone-product rule recorded under `~/.claude/projects/-Users-emman-GalileoOS/memory/` so future sessions don't re-derive them.

Open items deferred to Week 1 (NOT pre-flight blockers because they require code/PRs):
- `CLAUDE.md` and `AGENTS.md` at the repo root — Week 1, PR #2 (after the bootstrap commit).
- CI requires passing `/review` — Week 1, PR #4 (GitHub Actions workflow).

Branch protection is **not** deferred — it is configured atomically with the bootstrap commit per §1.5 below.

## 1.5 Pre-commit procedure (redline 5: atomic branch protection)

> **Live-state cross-reference.** The `required_approving_review_count` parameter in the JSON below is the value at protection-enable time. The live value was lowered from `1` to `0` on 2026-05-12 to break the single-account self-approval deadlock; see [`docs/decisions/0002-protection-approval-relaxation.md`](../decisions/0002-protection-approval-relaxation.md) for the rationale, consequences, and four reversal triggers. All other rules in this §1.5 JSON remain enforced live exactly as written.

The bootstrap commit lands on a branch that is **already protected** before any commit is pushed to it. There is no window where `main` is unprotected. Procedure is reproducible — anyone with admin on `MannyAmah/GalileoOS` can re-run it.

**Step 1 — Configure branch protection on `main` BEFORE any push.** The repo already exists with `main` as default branch (only `LICENSE` on it), so we apply protection to the existing `main` rather than creating it fresh:

```bash
gh api -X PUT repos/MannyAmah/GalileoOS/branches/main/protection \
  -H "Accept: application/vnd.github+json" \
  --input - <<'EOF'
{
  "required_status_checks": null,
  "enforce_admins": true,
  "required_pull_request_reviews": {
    "required_approving_review_count": 1,
    "dismiss_stale_reviews": true,
    "require_code_owner_reviews": false,
    "require_last_push_approval": false
  },
  "restrictions": null,
  "required_linear_history": false,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "block_creations": false,
  "required_conversation_resolution": true,
  "lock_branch": false,
  "allow_fork_syncing": false
}
EOF
```

Rules enforced after this call:
- **Required pull request reviews**: 1 approving review, stale reviews dismissed on new commits, conversation resolution required.
- **No force push**: `allow_force_pushes: false`.
- **No deletion**: `allow_deletions: false`.
- **Applied to administrators**: `enforce_admins: true` — Emmanuel and any future admin cannot bypass.
- **Required status checks**: `null` (intentionally empty) for the bootstrap PR because no CI exists yet. The required-checks list is **expanded** when PR #4 (Week 1) wires GitHub Actions, via a follow-up `PUT` to the same endpoint that names the new check contexts.

**Step 2 — Verify protection is live.** Before pushing anything:

```bash
gh api repos/MannyAmah/GalileoOS/branches/main/protection \
  --jq '{force_push: .allow_force_pushes.enabled, deletions: .allow_deletions.enabled, admins: .enforce_admins.enabled, reviews: .required_pull_request_reviews.required_approving_review_count}'
```

Expected: `{"force_push": false, "deletions": false, "admins": true, "reviews": 1}`.

**Step 3 — Create the bootstrap branch FROM `origin/main`.** The local tree is already initialized and pointing at `origin/main` (with the existing `LICENSE`); we branch off it:

```bash
git checkout -b bootstrap-stage0-plan
git add LICENSE .gitignore docs/galileo_os_infrastructure_plan.md docs/plans/STAGE_0_PLAN.md
git commit -m "chore(stage0): bootstrap repo with plan + STAGE_0_PLAN draft"
git push -u origin bootstrap-stage0-plan
```

`LICENSE` is included in the `git add` even though it's identical to `origin/main`'s copy — this is harmless (no actual change in the diff) and keeps the explicit list of "files in the bootstrap commit" complete.

**Step 4 — Open PR #1 against protected `main`.** No force-push, no direct commit to `main`. The PR is the only path:

```bash
gh pr create --base main --head bootstrap-stage0-plan --title "..." --body "..."
```

**Step 5 — Emmanuel reviews and approves; merge happens via PR review.** Squash merge preferred so `main` history shows one commit per logical change. Direct admin-bypass merge is structurally impossible because `enforce_admins: true`.

**Step 6 — After PR #1 merges, expand required status checks list as CI is wired in PR #4 (Week 1):**

```bash
gh api -X PATCH repos/MannyAmah/GalileoOS/branches/main/protection/required_status_checks \
  -f strict=true \
  -F 'contexts[]=lint' \
  -F 'contexts[]=type-check' \
  -F 'contexts[]=test' \
  -F 'contexts[]=dep-scan' \
  -F 'contexts[]=build-matrix'
```

This is the only step that adds checks; it is appended to PR #4's exit checklist.

**Why this works:** `main` becomes protected before any commit lands on it via push. The bootstrap PR cannot skip review, cannot be force-pushed, and cannot be bypassed by an administrator. The earlier `AskUserQuestion` answer that authorized "force-push as the initial commit" is **superseded by this redline 5** — no force-push happens. The user-chosen option (`git init` + remote + use existing `main` baseline) still holds; only the push mechanism changes from force-push to PR-through-protected-main.

## 2. Stage 0 success contract (what must be true on day 28)

Reproduced verbatim from `docs/galileo_os_infrastructure_plan.md` §6 / Stage 0 gate, plus the kickoff:

1. The repo builds, tests pass, `make up` brings the stack live on a fresh Ubuntu 24.04 VM.
2. `docs/closeouts/PROBE_MIRAGE_STAGE0.md` exists, is committed, and is the authoritative input for Layer 3 architecture.
3. The Hello Agent demo runs end-to-end with full Opik traces and cost attribution.
4. Three internal team members spin up a fresh Galileo, register a tenant, set a $5 monthly budget, run the demo agent 100× without anyone looking at logs. **Cost dashboard agrees with the Stripe metered-billing event count to the cent.**
5. A 30-minute install walkthrough video exists and a senior engineer who has never seen Galileo completes it end-to-end without help.

When all five are true, the Stage 1 plan PR opens. Not before.

## 3. Pre-registered failure handling

Per v7 rules 3 and 9, every observable failure in Stage 0 produces a closeout artifact in `docs/closeouts/`, named after the failing item, naming the **structural finding** (not "we ran out of time"). The Stage 0 gate is **not** relaxed to ship. If a closeout names a finding that invalidates an assumption in `docs/galileo_os_infrastructure_plan.md`, a `plan-deviation` issue is opened on GitHub and Stage 1 sequencing waits for sign-off on the revision.

Specific pre-registered failures Stage 0 must be prepared for:

| Failure | Detection | Closeout | Downstream consequence |
| --- | --- | --- | --- |
| Mirage layer-relocation discipline (no live probe required) | Week 2 closeout round | `CLOSEOUT_LAYER3_MIRAGE_RECONSIDERED.md` + canonical plan edits + `0003-mirage-layer-relocation.md` | First plan-deviation in the project. Mirage placed at Layer 5 (agent-side library) per Reading 2; the Layer 3 placement was a docs-discoverable mismatch, not a probe outcome. PR #10's apparatus is retained as a general kernel-side connector harness (rename in follow-up PR). |
| Temporal operational complexity exceeds capacity | Week 3 kernel boot | `CLOSEOUT_TEMPORAL_OPS_S0.md` | Stage 0 ships with `temporal server start-dev` only; Stage 2 promotion to Helm gets a dedicated spike before adoption |
| LiteLLM hop adds >50ms p95 latency | Week 3 benchmark | `CLOSEOUT_LITELLM_LATENCY_S0.md` | Deploy LiteLLM as sidecar to `galileo-agent-runner` instead of separate service |
| Cost meter disagrees with Stripe event count | Week 4 gate test | `CLOSEOUT_COST_METER_S0.md` | Stage 0 gate fails. Stage 1 does not begin. Root cause traced before retry. |
| Onboarding Crew stubs fail to produce a manifest on internal test workspace | Week 4 | `CLOSEOUT_ONBOARDING_S0.md` | Stub redesigned; Stage 1 Onboarding-Crew GA spec is revised before any implementation |

## 4. Weekly sequence

### Week 1 — Repo and CI/CD skeleton (2026-05-12 to 2026-05-19)

**Goal:** `make up` brings up the stack against empty service skeletons; CI green on a no-op PR; branch protection on; CLAUDE.md/AGENTS.md committed; ADR for the repo namespace deviation committed.

**Deliverables** (each shipped as its own PR; maker/checker on every merge). GitHub-PR numbers shown in parens where they differ from plan-PR labels (see [`docs/decisions/0002-protection-approval-relaxation.md`](../decisions/0002-protection-approval-relaxation.md) for the protection-relaxation context that produced the offset):

1. **plan-PR #1 (GitHub #1) — Bootstrap.** Initial commit: LICENSE (retained from current `main`), `docs/galileo_os_infrastructure_plan.md`, `docs/plans/STAGE_0_PLAN.md` (this file), `.gitignore`. Pushed via the §1.5 pre-commit procedure to **already-protected** `main` — no force-push, no direct commit to `main`. The PR is the only path. Branch protection is enabled before this PR is opened, not after it lands. **MERGED 2026-05-12.**
2. **plan-PR #2 (GitHub #3) — Operating discipline.** `CLAUDE.md` and `AGENTS.md` at repo root, codifying the v7 nine rules, the destructive-action definition, the read-only-by-default rule, the language boundaries from plan §5, and the link to this Stage 0 plan. Plus `docs/decisions/0001-repo-namespace.md` recording the `MannyAmah/GalileoOS` deviation and rationale. **MERGED 2026-05-12.** (GitHub #2 was auto-closed when the bootstrap branch was deleted at PR #1 merge; re-opened as GitHub #3 with byte-identical content.)
2a. **ADR-0002 micro-PR (GitHub #4) — Protection-approval relaxation.** Not originally in the plan; opened to document the protection-policy ADR after lowering `required_approving_review_count` from 1 to 0 to break the single-account self-approval deadlock. **MERGED 2026-05-12.** See [`docs/decisions/0002-protection-approval-relaxation.md`](../decisions/0002-protection-approval-relaxation.md).
3. **plan-PR #3 (GitHub #5) — Monorepo skeletons + protobuf v1 + minimal CI.** Originally scoped as skeletons-only (with CI deferred to plan-PR #4), but the reviewer required CI green on the empty skeletons before merge, so the minimal CI lane moved forward into this PR. **MERGED 2026-05-12.** Per plan §5.2, adapted to the kickoff's directory names:
   ```
   kernel/                Go — Temporal workers, agent-runner, gateway (skeleton main.go each)
     gateway/
     agent-runner/
     workflows/
   agents/                Python — LangGraph/CrewAI/Agno workers, ingestion
     onboarding/          Connector + Crawler stubs (Week 4)
     hello/               Hello Agent (Week 3)
   web/                   Next.js 16 admin (skeleton page, empty admin shell)
   mobile/                Expo (placeholder — no code in Stage 0)
   desktop/               Tauri (placeholder — no code in Stage 0)
   skills/                SKILL.md packs (empty in Stage 0)
   mcp-servers/           Custom Galileo MCP servers, empty in Stage 0
   schemas/               Protobuf contracts (one .proto: TaskInput, TaskResult, AgentOutput)
   deploy/
     compose/             docker-compose.yml from Appendix B
     helm/                placeholder
   docs/
     plans/  closeouts/  solutions/  decisions/
   .devcontainer/         Devcontainer (Go 1.23, Python 3.12, Node 22, Rust stable)
   Makefile               make up / test / lint / probe targets
   ```
   The kickoff says `agents/` and `web/`; plan §5.2 says `apps/` and `services/`. **Adopting kickoff names** since the kickoff is the more recent authoritative source for the repo layout.
4. **plan-PR #4 (GitHub #6, next) — CI expansion + status-check enforcement + devcontainer + Makefile.** Combines what the original plan called PR #4 (CI) and PR #5 (devcontainer + Makefile), minus the minimal CI that already shipped in plan-PR #3. Scope:
   - **CI expansion** on top of the minimal `go / python / web / protobuf` jobs:
     - `lint` (deeper than the current `vet`/`ruff`): golangci-lint (Go), ruff + black-check (Python), eslint + prettier (TS), rustfmt + clippy (Rust placeholder).
     - `dep-scan`: `govulncheck` (Go), `pip-audit` (Python), `npm audit --omit=dev` (Node).
     - Test wiring: keep `go test ./...` from plan-PR #3; add `pytest agents/` and `vitest run` (web). All allowed to be no-op; the workflow is the deliverable, not the test count.
     - `buf breaking` guard removal from `protobuf` job (carry-over from plan-PR #3 per the [carryover-commitments](../decisions/0002-protection-approval-relaxation.md) discipline — explicit checkbox in plan-PR #4's PR description before opening for review).
   - **Status-check enforcement on `main`** (already wired for the four PR #5 contexts; this step extends `required_status_checks` to add the new contexts above).
   - **`.devcontainer/devcontainer.json`** pinning Go 1.23, Python 3.12+, Node 22, Rust stable. Single multi-language image (mitigation noted below).
   - **`Makefile`** with `make up` (compose from Appendix B, placeholder service images until Week 3), `make test`, `make lint`, `make probe` (Week 2 placeholder).

**Week 1 exit criterion:** `make up` succeeds on a fresh Ubuntu 24.04 VM and brings up the Appendix B services (Postgres, Temporal, NATS, LiteLLM, Opik, Ollama). Galileo services are placeholder images (e.g., `nginx:alpine` standins) but the compose graph is intact. CI green on plan-PR #4 / GitHub #6. Branch protection enabled with the expanded required-status-check list. No PRs merged without review.

**Week 1 risks:**
- Devcontainer image sprawl across four languages may exceed Codespaces / local Docker budgets. Mitigation: use a single multi-language image rather than nested `.devcontainer/`s.
- Temporal-Postgres connection wiring in Appendix B is the most fragile piece. Mitigation: probe it as soon as Postgres + Temporal are up; if the official `auto-setup` image breaks, fall back to running `temporal server start-dev` standalone for Stage 0 (mitigation already named in plan).

### Week 2 — Mirage layer-relocation closeout (2026-05-19 to 2026-05-26)

**Revised scope.** The original Week 2 plan called for a live Mirage probe gating adoption at Layer 3. During PR #13's inline planning round (2026-05-13), reading Mirage's deployment-model documentation directly surfaced a structural mismatch: Mirage ships Python and TypeScript SDKs only, with no Go SDK or native server. Placing Mirage at Layer 3 (Galileo's Go kernel) would have required a permanent Python sidecar not named in the original plan. **The live probe is no longer the right experiment** — the placement decision is structural, not measurement-driven. Week 2 instead delivers the **first plan-deviation in the project's history**, following the four-part deviation template (closeout + canonical-plan edits + ADR + follow-up code change).

**Deliverables (Mon–Fri):**

1. `docs/closeouts/CLOSEOUT_LAYER3_MIRAGE_RECONSIDERED.md` — structural finding, three readings considered, choice (Reading 2 — Layer 5 agent-side), maker-checker provenance, downstream consequences, reversal triggers, v7-rule-3 framing, template-for-future-deviations.
2. Edits to `docs/galileo_os_infrastructure_plan.md` reflecting Mirage at Layer 5: §4.4 (Layer 3 substrate) Mirage row removed; §4.6 (Layer 5 integrations) Mirage reframed as agent-side library; §3.3 step 4 (Mount) reworded to clarify mounting is agent-side; destructive-action defense #3 reworded — kernel enforces *existence* of pre-write snapshot artifact, agents *produce* it.
3. `docs/decisions/0003-mirage-layer-relocation.md` — ADR with metadata, supersession info, named reversal triggers (Mirage publishes Go SDK / native server mode / Galileo kernel acquires need to crawl heterogeneous backends itself).
4. Edits to this file (`STAGE_0_PLAN.md`) reflecting the same: Week 2 scope revised, redline-4 row updated, Week 4 Onboarding Crew "if Mirage probe passed/failed" branching collapsed into per-agent choice.

PR #10's apparatus is **retained as a general kernel-side connector verification harness** rather than a Mirage-specific probe. The follow-up rename moved `kernel/probe/mirage/` → `kernel/probe/connector/` and updated the package name accordingly; the apparatus contents are unchanged.

**Calibration (Fri morning):** Before declaring Week 2 done, re-confirm that all four artifacts (closeout + canonical plan edits + ADR + STAGE_0_PLAN.md edits) are committed and internally consistent (each names the other; no stale "probe" references survive in any of them).

**Week 2 exit criterion:** `CLOSEOUT_LAYER3_MIRAGE_RECONSIDERED.md` + canonical plan edits + ADR-0003 + this file's edits all committed to `main` (single PR #13). Week 3 proceeds with the Onboarding Crew scaffolding under Reading 2: agents may import `mirage-ai` in-process or use discrete connector clients; the choice is per-agent.

**Week 2 risks:**
- Reading 2 implementation surface lands in agent code (Python). Risk surfaces during Week 4 when Onboarding Crew stubs are scaffolded, not during Week 2 itself.
- Reversal-trigger drift: if Mirage publishes a Go SDK during Stage 1+, ADR-0003's triggers fire and the placement decision is revisited. Until then, kernel does not import Mirage.

**Cold-engineer identification:** Engineer identified 2026-05-13 (per `CLAUDE.md` §Stage 0 process notes). Walkthrough scheduling happens in Week 4 when the Onboarding Crew scaffolding lands and the install walkthrough is ready to run end-to-end.

### Week 3 — Kernel boot (2026-05-26 to 2026-06-02)

**Goal:** Hello Agent demo runs 100× from a clean tenant under a $5 budget cap, Opik shows 100 traces, cost dashboard matches Stripe event count to the cent.

**Deliverables:**

1. **Compose stack live.** Appendix B's docker-compose.yml end-to-end on a Hetzner CCX23 (or equivalent local VM). All services healthy. Hostnames, ports, env vars per the appendix verbatim.
2. **`galileo-gateway` (Go).** Wraps LiteLLM with: tenant resolver (extracts tenant from signed JWT), per-tenant budget cap enforcement (reads tenant budget from Postgres, denies if month-to-date spend ≥ cap), request logging into Opik, signature verification on inbound webhooks. ~300 LOC; this is intentionally minimal.
3. **`galileo-agent-runner` (Go).** Connects to Temporal, registers one workflow (`HelloAgentWorkflow`) and one activity (`CallLLMActivity`). Activity calls the gateway, gateway calls LiteLLM, response returned. Cost metadata flows back into the workflow result. ~500 LOC.
4. **`galileo-web` (Next.js).** Single page at `:3001` that triggers the Hello Agent workflow, polls for completion, displays the response + the cost. ~150 LOC.
5. **Cost meter wiring.** LiteLLM emits per-tenant `usage` events. `galileo-gateway` aggregates them into a `cost_events` Postgres table with the schema `(tenant_id UUID, event_ts TIMESTAMPTZ, cost_cents BIGINT, provider TEXT, model TEXT, request_id TEXT PRIMARY KEY)`. Cost is stored as integer cents — no floating-point dollars. A nightly cron (Temporal workflow) reconciles `cost_events` against Stripe metered-billing usage records for the same tenant. Discrepancy logged as a critical alert.
6. **Reconciliation script (gate-test method).** A Go program at **`kernel/cost-recon/recon.go`** runs at the end of the Stage 0 gate test. Behavior:
   - **Inputs:** `--tenant-id`, `--window-start` (inclusive, RFC 3339), `--window-end` (exclusive, RFC 3339). Window is `[window_start, window_end)` — start included, end excluded. This avoids double-counting events at boundary timestamps.
   - **Step 1:** `SELECT request_id, event_ts, cost_cents FROM cost_events WHERE tenant_id = $1 AND event_ts >= $2 AND event_ts < $3 ORDER BY event_ts, request_id`.
   - **Step 2:** Fetch Stripe metered-billing usage records for the same `tenant_id` (mapped to the Stripe subscription item ID) over the same window via the Stripe API; pull `(timestamp, quantity)` pairs where `quantity` is cents-as-integer.
   - **Step 3:** Join the two result sets on `(tenant_id, event_timestamp)`. The join key uses second-precision timestamps; sub-second clock skew is normalized by rounding both sides to the nearest second.
   - **Step 4:** For each joined row, assert `cost_events.cost_cents == stripe.quantity` as integer equality. Any row where the integer values differ is a failure. Any row present on one side but missing on the other is a failure (i.e., the joined-row count must equal both source counts).
   - **Step 5:** Exit code `0` on a clean run; non-zero on any discrepancy. **A non-zero exit code is a Stage 0 gate failure regardless of magnitude** — one-cent diffs fail the gate. The output writes the offending rows to `docs/closeouts/COST_RECON_<date>.json` for the closeout doc.
   - The script is invoked by `make stage0-gate` (added in PR #5 of Week 1) and runs as the final assertion of the gate test in Week 4.
7. **Opik integration.** Every gateway request emits a span. Every workflow execution shows as one parent span with N LLM children. Stage 0 acceptance test runs the demo 100× and confirms 100 parent spans in Opik.

**Week 3 exit criterion:** From a fresh Ubuntu VM, `make up && make demo` registers a test tenant, sets $5 budget, runs Hello Agent 100×, exits 0. Opik shows 100 traces. Cost dashboard total matches the sum of Stripe `usage_record` POST bodies for the test tenant.

**Week 3 risks:**
- LiteLLM proxy adds latency; if p95 exceeds 50ms hop, sidecar deployment (mitigation noted in §3 above).
- Cost-meter / Stripe reconciliation is the highest-risk piece — fundamental to gate test 4. Mitigation: write the reconciliation cron as the first deliverable of Week 3, not the last.

### Week 4 — Onboarding Crew scaffold + Stage 0 gate (2026-06-02 to 2026-06-09)

**Goal:** Connector + Crawler agent stubs produce a manifest against an internal test workspace; Stage 0 gate test runs and passes.

**Internal test workspace:**
- A small GitHub repo (≤50 files, all in `MannyAmah` org or a throwaway). Read-only PAT.
- A small Slack export (last 30 days from the test workspace).
- A small Google Drive folder (≤20 files, shared with the workspace test account).

**Deliverables:**

1. **`agents/onboarding/connector.py`.** Python LangGraph agent that authenticates each source and writes the per-source credentials into the tenant's Infisical bucket. Downstream agents (Crawler, Ingestion) may import `mirage-ai` and mount each authenticated source in-process, or instantiate per-source MCP clients from the discrete fallback set named below. Per-agent choice. **No ingestion** in Stage 0 — Stage 1's job.
2. **`agents/onboarding/crawler.py`.** Walks every connected source and emits a manifest (list of paths, sizes, hashes, content types) into a `tenant_manifests` Postgres table. Caps per plan §3.3 step 5 (50K docs / 6h wall clock / $50 LLM spend) enforced as hard limits.
3. **Manifest validator.** A simple Go binary in `kernel/manifest-check/` that reads the manifest, computes expected vs. actual counts per source, and fails CI if a regression is detected on the test workspace.
4. **30-minute install walkthrough video.** Recorded by Emmanuel (or by a designated maker, with the QA agent doing the run). Hosted on the eventual `galileoos.com` or, in Stage 0, on YouTube as Unlisted with a link checked into `docs/onboarding/install_walkthrough.md`.
5. **Stage 0 gate test.** Three internal teammates (Emmanuel + two volunteers) each spin up a fresh instance, run through the install walkthrough, register a tenant, set $5 budget, run Hello Agent 100×. Their results are committed to `docs/closeouts/STAGE_0_GATE.md` with names, dates, and any deviations.

**Under Reading 2 (Mirage at Layer 5, agent-side):** the Onboarding Crew scaffolding in Week 4 lands with Python agents that may import `mirage-ai` in-process or call discrete MCP servers directly — the choice is per-agent. The three discrete MCP servers for the internal test workspace are **locked at sign-off of this plan**, available to both Mirage-using and Mirage-bypassing agents:

- **GitHub:** `@modelcontextprotocol/server-github` — Anthropic's reference MCP server in the curated `punkpeye/awesome-mcp-servers` catalog (plan §4.6 / Dev/Ops row). MIT-licensed, Node.js, deployed as a sandboxed subprocess by `agents/onboarding/connector.py`. Read-only PAT scopes only (`contents:read`, `metadata:read`).
- **Slack:** `@modelcontextprotocol/server-slack` — Anthropic's reference MCP server in the curated catalog (plan §4.6 / Communication row). MIT-licensed, Node.js. Read-only OAuth bot scopes only (`channels:history`, `channels:read`, `files:read`, `users:read`).
- **Google Drive:** `@modelcontextprotocol/server-gdrive` — Anthropic's reference MCP server in the curated catalog (plan §4.6 / Files/Docs row). MIT-licensed, Node.js. Read-only OAuth scope (`drive.readonly`).

All three live under the official `modelcontextprotocol/servers` repository, vendored at a pinned commit at `mcp-servers/fallback/` if needed. No write scopes — per the locked decision in `docs/galileo_os_infrastructure_plan.md` §kickoff / read-only-by-default.

**Week 4 exit criterion (== Stage 0 gate):**
1. ✅ `make up` works on a fresh Ubuntu 24.04 VM (verified by the three teammates).
2. ✅ `CLOSEOUT_LAYER3_MIRAGE_RECONSIDERED.md` + canonical plan edits + ADR-0003 committed (from Week 2).
3. ✅ Hello Agent demo: 100 runs, 100 Opik traces, cost = Stripe to the cent (verified by all three teammates).
4. ✅ Onboarding Crew stubs produce a valid manifest against the test workspace.
5. ✅ Walkthrough video exists; one of the three teammates ran the install end-to-end without help.

If any item fails, `docs/closeouts/CLOSEOUT_STAGE0.md` is written naming the structural finding. Stage 1 plan PR does **not** open.

**Week 4 risks:**
- "Senior engineer who has never seen Galileo completes the install" is hard to source from a 3-person internal team. Mitigation: pre-register one outside engineer (paid hourly) as the cold tester; they sign an NDA, do the install, write a short note for the closeout.
- Manifest schema collisions between Mirage-using agents and Mirage-bypassing agents. Mitigation: schema designed to be union of both modes; `source_kind` field discriminates.

## 5. Maker / checker assignments (v7 rule 5)

Stage 0 is small — likely just Emmanuel plus Claude sessions. The discipline still applies:

- **Maker** is the agent (Claude Opus 4.7) authoring code or docs.
- **Checker** is **never** the same agent in the same session. For Stage 0:
  - PRs go up; Emmanuel reviews and approves before merge.
  - For UI/Hello-Agent flow, `/qa` is run against the staging URL by a separate Claude session (or by Emmanuel manually) and the screenshot evidence committed.
  - The Week 2 Mirage layer-relocation closeout, canonical plan edits, and ADR-0003 are reviewed by Emmanuel before the PR is merged. The maker-checker iteration runs artifact-by-artifact (closeout, then plan edits, then ADR) rather than reviewing all three at once.

No exceptions. Even a typo fix in a Stage 0 doc is a PR.

## 6. Compounding (v7 rule 7)

Every non-trivial finding during Stage 0 is captured in `docs/solutions/<topic>.md`:

- Devcontainer assembly across four languages → `docs/solutions/multi-language-devcontainer.md`
- Temporal-Postgres wiring gotchas → `docs/solutions/temporal-postgres-wiring.md`
- LiteLLM tenant context propagation → `docs/solutions/litellm-tenant-context.md`
- Cost meter ↔ Stripe reconciliation → `docs/solutions/cost-meter-stripe-recon.md`
- Mirage layer-relocation lessons → integrated into `CLOSEOUT_LAYER3_MIRAGE_RECONSIDERED.md` + the "read the dependency's deployment-model documentation before encoding it in the plan" discipline pattern compounds for future vendor evaluations.

Future sessions read `docs/solutions/` before starting work. The cost is 10 minutes per finding; the value compounds for years.

## 7. Stage 0 success artifacts (recap)

When Stage 0 closes:

1. `MannyAmah/GalileoOS` is a buildable, testable monorepo with green CI on `main`, branch protection on, and the eight bare service skeletons.
2. `docs/closeouts/PROBE_MIRAGE_STAGE0.md` is committed and is the authoritative input for Layer 3.
3. Hello Agent runs durably with full Opik traces and to-the-cent cost attribution.
4. `docs/closeouts/STAGE_0_GATE.md` documents the three-teammate gate test outcome and the cold-engineer walkthrough.
5. (Conditional) `docs/solutions/*.md` holding the lessons compounded across Weeks 1–4.

When all five exist and items 1–4 pass, the **Stage 1 plan PR** opens. Until then, no Stage 1 work begins.

## 8. Escalation

Per kickoff: if anything in `docs/galileo_os_infrastructure_plan.md` turns out to be wrong, do **not** silently work around it. Open an issue tagged `plan-deviation`, name the structural finding, propose the revised approach, wait for Emmanuel's sign-off. Plan deviations are normal; silent workarounds are not.

Examples that would trigger a `plan-deviation` issue:
- A vendor-evaluation read surfaces a structural mismatch with the plan's role for that vendor (e.g., the Mirage layer-relocation deviation in PR #13 — closeout + canonical-plan edits + ADR + follow-up code change; first deviation in the project's history, sets the four-part template for future ones).
- Appendix B docker-compose images turn out to be unavailable or incompatible with Ubuntu 24.04.
- LiteLLM's tenant-context model can't actually attribute cost to tenants — would change §2.3's claim that "no homegrown meter" is needed.

## 9. Sign-off

This plan goes from DRAFT to APPROVED when Emmanuel explicitly says so in conversation (or comments approval on the PR if we open one for it). Week 1 work begins only after APPROVED. The current status is **DRAFT — awaiting sign-off**.

---

*This document is part of the Galileo OS planning corpus. It is not the spec — the spec is `docs/galileo_os_infrastructure_plan.md`. This document is how Stage 0 executes against that spec without softening any of its gates.*
