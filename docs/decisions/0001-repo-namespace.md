# ADR-0001 — Repo namespace: `MannyAmah/GalileoOS` instead of `galileoos/galileo-os`

| Field | Value |
| --- | --- |
| **Status** | Accepted |
| **Date** | 2026-05-12 |
| **Decider** | Emmanuel (founder) |
| **Author** | Claude Opus 4.7 (1M context) under Emmanuel's direction |
| **Plan deviation** | Yes — supersedes the "Locked decisions" list in the project kickoff and the `galileoos/galileo-os` references in [`docs/galileo_os_infrastructure_plan.md`](../galileo_os_infrastructure_plan.md) §6.0 |

## Context

The project kickoff and the plan locked the canonical repo at `github.com/galileoos/galileo-os`. The actual repo created for Stage 0 work lives at [`github.com/MannyAmah/GalileoOS`](https://github.com/MannyAmah/GalileoOS). The discrepancy was flagged as a plan deviation at pre-flight (2026-05-12) per the rule "do not silently work around plan defects."

The plan's namespace choice affects:

- The `curl|bash` installer URL on `galileoos.com` (one-line install path).
- The Helm chart repository URL.
- Every `git clone` line in the docs.
- The GitHub OAuth app's redirect URI and audience claim.
- The brand surface ("Galileo OS is at galileoos/galileo-os").

A silent rename later costs a 301 redirect setup, a Helm chart re-publish, OAuth app re-registration, and a docs sweep. It is not free.

## Options considered

1. **Use `MannyAmah/GalileoOS` as-is.** Cheapest path. Personal-account repo is canonical for the foreseeable future. If a `galileoos` org is created later, transfer the repo and add a 301 redirect.
2. **Create `galileoos` GitHub org now, transfer + rename to `galileo-os`.** Matches the plan exactly. Costs ~15 minutes of human time to create the org, transfer the repo, set defaults, invite collaborators.
3. **Hybrid: keep `MannyAmah/GalileoOS` as canonical for Stage 0, plan org migration for Stage 1.** Defers the choice.

## Decision

**Option 1 — use `MannyAmah/GalileoOS` as-is.**

Decision recorded in conversation on 2026-05-12. The personal-account repo is canonical for Stage 0 and beyond unless explicitly migrated. The plan's `galileoos/galileo-os` reference is **superseded** by this ADR for all distribution-surface artifacts (installer URL, Helm repo URL, docs, OAuth app).

## Consequences

**Positive:**

- Zero time spent on repo migration in Stage 0.
- The repo URL is stable from the first commit; no broken links during Stage 0.
- Emmanuel retains direct admin control (his GitHub identity owns the repo).

**Negative / risks:**

- The `galileoos.com` brand surface and the `MannyAmah/GalileoOS` source surface are split — a customer who Googles "galileo os github" may not find the personal-account repo as easily as an org-namespaced one.
- If Galileo OS hires its first employee in Stage 1, transferring the repo into an org becomes a one-way migration that breaks any pre-existing `git clone` lines in customer self-hosts. Mitigation: when migration happens, GitHub auto-creates a 301 redirect for the repo URL and most clone lines continue working.
- The plan and the README will reference `MannyAmah/GalileoOS`; if migrated later, the plan needs a follow-up ADR-000N reversing this one.

## Triggers for revisiting

This ADR should be revisited (and likely reversed) when any of:

- Galileo OS hires its first non-Emmanuel engineer.
- The Stage 0 gate passes and the `galileoos.com` installer URL goes live to external users.
- A vendor or compliance flow requires an org-namespaced GitHub repo (e.g., GitHub Marketplace listing for a Galileo MCP server).
- Emmanuel decides to open-source the repo to a community of external contributors.

When revisited: open ADR-000N (supersedes 0001), perform the org transfer, update the plan's distribution-surface references in a separate PR.

## References

- [`docs/galileo_os_infrastructure_plan.md`](../galileo_os_infrastructure_plan.md) §6.0 — original namespace lock.
- Project kickoff message (2026-05-12) — "Locked decisions" list naming `galileoos/galileo-os`.
- AskUserQuestion response (2026-05-12) — Emmanuel selected "Use MannyAmah/GalileoOS as-is."
- [`docs/plans/STAGE_0_PLAN.md`](../plans/STAGE_0_PLAN.md) §1 — references this ADR.
