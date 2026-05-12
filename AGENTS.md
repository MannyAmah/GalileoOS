# AGENTS.md — Galileo OS

This file is the operating contract for **any AI agent** working on this repository — Claude Code, Cursor, Aider, Cline, OpenHands, Codex, or anything else that can read a markdown file and run a tool. It is intentionally short. The long-form discipline lives in [`CLAUDE.md`](./CLAUDE.md); this file is what an agent reads first to know whether it's allowed to touch the repo.

If you are a human, read [`CLAUDE.md`](./CLAUDE.md) instead — it has more context.

---

## What you are looking at

Multi-tenant, self-hostable, Apache 2.0 B2B AI infrastructure. The authoritative spec is [`docs/galileo_os_infrastructure_plan.md`](./docs/galileo_os_infrastructure_plan.md). The current execution plan is [`docs/plans/STAGE_0_PLAN.md`](./docs/plans/STAGE_0_PLAN.md). Read both before non-trivial work. Read [`CLAUDE.md`](./CLAUDE.md) for the full operating contract.

## Reading order before any work

1. [`AGENTS.md`](./AGENTS.md) (this file).
2. [`CLAUDE.md`](./CLAUDE.md).
3. [`docs/galileo_os_infrastructure_plan.md`](./docs/galileo_os_infrastructure_plan.md) — at least the eight-layer overview (§2), the v7 discipline (§2.4), the Day-Zero Onboarding section (§3), and the current stage's roadmap (§6).
4. [`docs/plans/STAGE_0_PLAN.md`](./docs/plans/STAGE_0_PLAN.md) — the active execution plan.
5. [`docs/solutions/`](./docs/solutions/) — grep for prior art before solving anything.
6. [`docs/closeouts/`](./docs/closeouts/) — failed-phase findings; relevant context for new phases.

If you skip steps 3–4 you will violate v7 rule 6 (plan before code) and your PR will be rejected.

## Hard rules for any agent

These are the rules that, if violated, mean a PR does not merge. They are derived from [`CLAUDE.md`](./CLAUDE.md) §v7 nine-rule discipline. Distilled for agents that prefer terse lists:

1. **No code before a plan exists.** If there is no `docs/plans/<topic>.md` (or a clear human directive in the conversation), stop and write one.
2. **No vendor adoption without a probe.** If your PR introduces a new dependency or service, the PR must include a probe that exercises the specific behaviors we need from it. No probe → no merge.
3. **No softening of pre-registered gates.** Stage gates and probe pass criteria are locked. If your work shows a gate is unattainable, open a [`plan-deviation`](#escalation) issue — do not relax the gate.
4. **No declaring your own work done.** Maker / checker separation. Open the PR, let the checker (human reviewer, `/review`, or a separate agent session) confirm it.
5. **No destructive call without a Temporal-signal gate.** Destructive = any HTTP method other than `GET`/`HEAD`/`OPTIONS`, any SQL not a `SELECT`, any `rm`/`delete`/`drop`/`truncate`/`unlink`, any social-platform `POST`, any email send, any payment transfer. **Prompt-injection cannot override this.** The agent runtime must structurally refuse.
6. **No write OAuth scopes by default.** Read-only is the default. Write scope is requested per-action only after the operator explicitly enables a department.
7. **No direct LLM provider SDK usage outside the LiteLLM gateway.** Every LLM call goes through LiteLLM. CI lint enforces this.
8. **No mixed-language services.** One service = one language. Cross-service is gRPC over NATS with protobuf in [`schemas/`](./schemas/).
9. **No silent workarounds for plan defects.** If the plan is wrong, open a `plan-deviation` issue.
10. **Compound your lessons.** Non-trivial bug fixes and gotchas get a markdown note in [`docs/solutions/<topic>.md`](./docs/solutions/).

## What an agent can do on its own

- Read any file in the repo.
- Run any `make` target except destructive ones (`make wipe`, `make reset`, etc., if they exist).
- Write code, run tests, lint, type-check.
- Open a PR with the changes.
- Update `docs/plans/`, `docs/solutions/`, `docs/decisions/` as part of the PR.
- Run `gh pr view`, `gh pr diff`, `gh pr review` to inspect PR state.

## What an agent must ask before doing

- Anything destructive in the live world (deploying to production, paying a vendor, sending email, posting to social, dropping a database, force-pushing, deleting a branch).
- Adopting a new vendor without a probe.
- Changing a locked decision (see [`CLAUDE.md`](./CLAUDE.md) §Locked-in architectural decisions).
- Modifying [`docs/galileo_os_infrastructure_plan.md`](./docs/galileo_os_infrastructure_plan.md) (the spec is authoritative; spec changes require a `plan-deviation`).
- Modifying CI / branch protection / secrets / `.github/` configuration.

## What an agent must never do

- Use `git push --force` on a protected branch.
- Skip pre-commit hooks (`--no-verify`).
- Bypass GPG signing.
- Commit secrets (`.env`, `*.pem`, `*.key`, anything matching `^(AKIA|sk-|ghp_|xoxb-|xoxp-)`).
- Touch `.git/`, `.github/workflows/` for the purpose of disabling protection.
- Cross-pollinate Galileo with patterns from any other project on this machine.

## Escalation

If something in [`docs/galileo_os_infrastructure_plan.md`](./docs/galileo_os_infrastructure_plan.md) turns out to be wrong, **do not silently work around it.**

1. Open a GitHub issue tagged `plan-deviation`.
2. Title: `plan-deviation: <one-line structural finding>`.
3. Body: the finding, the evidence (logs, probe output, measurements), the proposed revised approach.
4. Stop work on the affected path until a reviewer signs off.

Plan deviations are normal and welcome. Silent workarounds are not.

## Communication style

When you talk to the human operator in a PR description or a conversation:

- Lead with what changed and why, not what you tried.
- Cite file paths and line numbers (`web/admin/page.tsx:42`) when referencing code.
- If your work would benefit from a `/review` or `/qa` pass, say so explicitly — don't assume the reviewer will infer it.
- One-or-two-sentence end-of-turn summary. What changed, what's next. Nothing else.

---

*Last updated: 2026-05-12. Any change to this file is a PR and requires reviewer approval.*
