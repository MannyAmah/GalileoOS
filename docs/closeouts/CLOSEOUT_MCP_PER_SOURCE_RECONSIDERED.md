# CLOSEOUT — Per-source MCP server set reconsidered (Stage 0 Onboarding Crew)

**Phase:** Stage 0 Week 4, Onboarding Crew scaffolding (deliverables 1–2).
**Outcome:** Substituted. Three reference MCP servers → per-source dispatch (Docker-subprocess MCP for github; direct SDK for slack and gdrive).
**Date:** 2026-05-16.
**Status:** Third plan-deviation in the project's history. Follows the four-part deviation template established by PR #13 (Mirage) and continued by PR #17 (Opik): closeout + canonical plan edits + ADR + the code that implements the substitution, all in one PR.

## Structural finding

The Week 4 inline-plan round's discovery pass surfaced two compounding structural findings against the three reference MCP servers locked at the original plan's sign-off.

**Finding 1 — Upstream archive.** Three of the four named servers (`@modelcontextprotocol/server-slack`, `@modelcontextprotocol/server-gdrive`, and the npm-distributed `@modelcontextprotocol/server-github`) were archived upstream on commit `d53d6cc75c` of `modelcontextprotocol/servers` on 2025-05-29 — five and a half months before Week 4 opened. Last working commits: Slack `52db0d9899` (2025-04-22), GDrive `96352032fc` (2025-05-06). The reference implementations are no longer maintained. The github reference server's archive predates the others; GitHub itself published a vendor-owned replacement.

**Finding 2 — Distribution-channel mismatch.** GitHub's vendor-maintained replacement at `github/github-mcp-server` (v1.0.4 released 2026-05-11, MIT, 29.8k stars) is **not** distributed via npm. The plan's `npx -y` invocation pattern was the wrong shape against the actual distribution model. The release assets are Go binaries (`github-mcp-server_Darwin_arm64.tar.gz`, `github-mcp-server_Linux_x86_64.tar.gz`, etc.) and a Docker image (`ghcr.io/github/github-mcp-server`). GitHub's published installation docs lead with the Docker invocation; the npm-distributed reference server the plan named simply does not exist in a non-deprecated form.

Slack and Google have not published vendor-maintained MCP replacements. Their respective Python SDKs (`slack_sdk`, `google-api-python-client`) are vendor-maintained, well-documented, and cover the read-only enumeration surface Stage 0 needs.

The mismatch is structural, not measurement-driven. No volume of probe traffic would re-animate the deprecated reference servers; no incremental tuning would change npm into Docker. The discovery pattern is the fourth instance of the documented dependency-shape pattern: see [`docs/solutions/SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md`](../solutions/SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md), which captures the axis decomposition (deployment topology, container topology, ecosystem utilities, **distribution channel**) added in PR-D.

## Why not relax-the-spec

Four options were considered before settling on per-source dispatch:

1. **Pin to last-working archived commits.** Rejected — three of three reference servers archived simultaneously is not a maintenance status the project can rely on for a Stage 0 gate. Pinning to a 2025-04-22 commit of `server-slack` and a 2025-05-06 commit of `server-gdrive` accumulates security and behavioral drift with no upstream to surface it.
2. **Wait for vendor-maintained MCP replacements from Slack and Google.** Rejected — no public roadmap commits either vendor to publishing an MCP server. Stage 0 has a Week 4 gate; deferring an indefinite amount of time on an indefinite signal is the wrong shape for a calendar-bound deliverable.
3. **All direct SDKs (drop MCP entirely for Stage 0).** Rejected — collapses the long-run framing where MCP is the composability surface for agent-side tool dispatch. GitHub's vendor-maintained MCP server exists and is exactly the kind of dependency the plan was right to lean on for tool exposure.
4. **Per-source dispatch: MCP where vendor-maintained, direct SDK where not.** **Chosen.** Connector and Crawler dispatch by `source_kind`. For github: Docker subprocess of `ghcr.io/github/github-mcp-server:v1.0.4` via the mcp Python SDK's `stdio_client`. For slack: `slack_sdk.WebClient`. For gdrive: `googleapiclient.discovery.build("drive", "v3", credentials=service_account.Credentials.from_service_account_file(...))`. The composability framing stays valid for github; the direct-SDK fallback for slack/gdrive is the same shape as any other read-only OAuth integration Galileo will eventually need.

Within option 4, three sub-options for the github invocation were considered:

- **Docker subprocess** (`docker run -i --rm --init -e GITHUB_PERSONAL_ACCESS_TOKEN ghcr.io/github/github-mcp-server:v1.0.4`). Adopted. Docker is already a dev-stack prerequisite (`make up`); zero new prerequisites for developers. The `--init` flag is a five-character addition that prevents zombie-container failures during Temporal workflow cancellation paths, where Python-side SIGTERM might otherwise reach a PID 1 unable to handle the signal cleanly.
- **Local Go binary.** Would have required adding either Go ≥1.26 (already present in dev stack but not yet a *runtime* prerequisite for non-Go developers) or a `make install-github-mcp-server` target that downloads the tarball for the host platform. Adds a binary-install step the Docker approach doesn't.
- **Remote HTTP MCP at `https://api.githubcopilot.com/mcp/`.** Rejected for Stage 0; named as the fourth reversal trigger in ADR-0005. The structural argument: Option A treats `github-mcp-server` like LiteLLM or Postgres — Galileo runs it locally. The remote-HTTP variant treats it like a vendor-hosted API — vendor runs it, Galileo calls it. For Stage 0 these are equivalent capability-wise, but the choice compounds into Stage 1+ operational posture (vendor uptime dependency, enterprise egress policies, GitHub-side incidents becoming Galileo runtime incidents). All three considerations lean toward keeping `github-mcp-server` symmetric with the other local runtime services.

## What ships in PR-D

| Artifact | Path | Purpose |
|---|---|---|
| Closeout (this file) | `docs/closeouts/CLOSEOUT_MCP_PER_SOURCE_RECONSIDERED.md` | Names the structural finding; required by v7 rule 3 for any phase or deliverable with a pre-registered gate, pass *or* fail. |
| ADR | `docs/decisions/0005-mcp-per-source-vs-mixed.md` | Locks in per-source dispatch with four reversal triggers. |
| Plan edits | `docs/plans/STAGE_0_PLAN.md` §Week 4 deliverables 1–2 | Replaces the three-reference-server list with the per-source dispatch; encodes Drift-7 (Opik → Jaeger wording), Drift-10 (`kernel/manifest-check/` → `kernel/cmd/manifest-check/`), and the §3.5 gate scoping note for Skill precision. |
| Code | `agents/onboarding/*.py`, `kernel/cmd/gateway/migrations/0003_tenant_credentials.sql`, `kernel/cmd/gateway/migrations/0004_tenant_manifests.sql`, `kernel/cmd/manifest-check/main.go`, CI YAML additions, walkthrough doc | Implements Connector + Crawler + Temporal worker + CLI + credentials store + manifest-check validator. Same gate-test contract holds. |
| Solutions-doc compound | `docs/solutions/SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md` | Fourth instance + axis decomposition (deployment topology / container topology / ecosystem utilities / distribution channel) + four per-axis verification commands. |

All artifacts land in PR-D (the single-commit-set discipline from PR #13 and PR #17 holds — the deviation is indivisible from the code that implements it).

## Reversal triggers

Per-source dispatch should be reconsidered if any of the following fire in Stage 1+:

1. **Slack publishes a vendor-maintained MCP server.** Migrate slack from `slack_sdk` to that server; preserve direct-SDK dispatch as the fallback shape if the vendor server has gaps.
2. **Google publishes a vendor-maintained MCP server.** Same as Slack.
3. **A sixth source-kind needs adding** and the per-source dispatch boilerplate becomes load-bearing. At that point, evaluate either (a) generalizing the dispatch to a registry-driven plugin pattern or (b) collapsing to all-MCP if the ecosystem has matured (vendor servers from each of the six).
4. **Docker becomes unavailable as a dev-stack prerequisite** (offline-first installs, regulated environments). Migrate github from local Docker subprocess to the remote HTTP MCP server at `https://api.githubcopilot.com/mcp/` via the mcp Python SDK's HTTP transport.

Until then, per-source dispatch handles Stage 0's Onboarding Crew. The reference-server set named in the original plan is the wrong shape for the deliverable.

## Cross-references

- Origin of the discovery: Week 4 inline plan, 2026-05-16, captured in this closeout.
- Pattern documentation: [`docs/solutions/SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md`](../solutions/SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md) — github-mcp-server is instance 4, contributes the distribution-channel axis.
- Deviation template precedents: [`docs/closeouts/CLOSEOUT_LAYER3_MIRAGE_RECONSIDERED.md`](CLOSEOUT_LAYER3_MIRAGE_RECONSIDERED.md) (2026-05-13), [`docs/closeouts/CLOSEOUT_OPIK_RECONSIDERED.md`](CLOSEOUT_OPIK_RECONSIDERED.md) (2026-05-14).
- ADR for this substitution: [`docs/decisions/0005-mcp-per-source-vs-mixed.md`](../decisions/0005-mcp-per-source-vs-mixed.md).
