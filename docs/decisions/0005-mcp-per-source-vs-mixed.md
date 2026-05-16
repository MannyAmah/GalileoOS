# ADR-0005 — Onboarding Crew connectors: per-source dispatch (MCP for github, direct SDK for slack and gdrive)

| Field | Value |
| --- | --- |
| **Status** | Accepted |
| **Date** | 2026-05-16 |
| **Decider** | Emmanuel (founder) |
| **Author** | Claude Opus 4.7 (1M context) under Emmanuel's direction |
| **Supersedes** | The three-reference-MCP-server set named in [`docs/plans/STAGE_0_PLAN.md`](../plans/STAGE_0_PLAN.md) §Week 4 (pre-PR-D revision) |
| **Plan deviation** | Yes — **third plan-deviation in the project's history.** Follows the four-part template from ADR-0003 (Mirage) and ADR-0004 (Opik): closeout + canonical plan edits + this ADR + the code that implements the substitution. |
| **Companion artifacts** | [`docs/closeouts/CLOSEOUT_MCP_PER_SOURCE_RECONSIDERED.md`](../closeouts/CLOSEOUT_MCP_PER_SOURCE_RECONSIDERED.md) — full structural finding. [`docs/solutions/SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md`](../solutions/SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md) — the dependency-shape pattern this finding instantiates (4th instance, contributes the distribution-channel axis). |

## Context

The original plan named three MCP servers at Week 4 sign-off — `@modelcontextprotocol/server-github`, `@modelcontextprotocol/server-slack`, `@modelcontextprotocol/server-gdrive` — under the framing that all three were Anthropic's reference servers in the curated `modelcontextprotocol/servers` repository, vendored at a pinned commit if needed.

PR-D's inline-plan discovery pass surfaced two compounding findings:

1. **All three reference servers were archived upstream on commit `d53d6cc75c` on 2025-05-29.** No upstream maintenance; pinning to last-working commits accumulates drift without a way to surface it.
2. **GitHub's vendor-maintained replacement (`github/github-mcp-server` v1.0.4) is not an npm package.** It distributes as a Docker image (`ghcr.io/github/github-mcp-server`) and Go binary release tarballs. The plan's `npx -y` invocation was the wrong shape against the actual distribution model. Slack and Google have not published vendor-maintained MCP servers.

Full structural reasoning, the four options considered, and why each was rejected or adopted: [`docs/closeouts/CLOSEOUT_MCP_PER_SOURCE_RECONSIDERED.md`](../closeouts/CLOSEOUT_MCP_PER_SOURCE_RECONSIDERED.md).

## Decision

The Onboarding Crew's Connector and Crawler agents **dispatch by `source_kind`**. For each source-kind, the integration mode is chosen by the actual upstream availability of a vendor-maintained MCP server, not by an a-priori uniformity preference.

| Source-kind | Integration mode | Invocation | Read-only scope set |
|---|---|---|---|
| `github` | MCP via Docker subprocess | `docker run -i --rm --init -e GITHUB_PERSONAL_ACCESS_TOKEN ghcr.io/github/github-mcp-server:v1.0.4` via `mcp.StdioServerParameters` + `stdio_client` | Fine-grained PAT: `contents:read`, `metadata:read` |
| `slack` | Direct SDK | `slack_sdk.WebClient(token=bot_token)` | Bot scopes: `channels:read`, `groups:read`, `users:read` |
| `gdrive` | Direct SDK | `googleapiclient.discovery.build("drive", "v3", credentials=service_account.Credentials.from_service_account_file(path, scopes=...))` | OAuth scope: `https://www.googleapis.com/auth/drive.readonly` |

**Why this shape.** GitHub maintains an actively-developed MCP server; Slack and Google do not. Per-source dispatch treats each integration at its actual upstream maturity, not at the lowest common denominator. The composability framing that motivated MCP for tool exposure stays valid where vendor-maintained MCP exists; the direct-SDK fallback is the same shape as any other read-only OAuth integration Galileo will eventually need.

**Why Docker subprocess over remote HTTP MCP for github.** Symmetry of operational posture. Docker subprocess treats `github-mcp-server` like LiteLLM, Postgres, Temporal — runtime services Galileo runs locally. Remote HTTP MCP at `https://api.githubcopilot.com/mcp/` would treat it like Stripe's metered-billing API — vendor-hosted, called over the public internet. Stage 0 has zero of the latter category among its core runtime dependencies; introducing a third category (vendor-hosted MCP we call but didn't plan around) adds operational surface area without earning it. The remote-HTTP variant is the fourth reversal trigger below if Docker availability changes.

**Why `--init`.** Without `--init`, a Python-side SIGTERM during Temporal workflow cancellation might reach a Docker container whose PID 1 process can't handle the signal cleanly. The `--init` flag installs a tini-equivalent reaper as PID 1, ensuring signal forwarding and child-process reaping. Standard pattern for short-lived subprocess containers; five characters; prevents zombie-container failures.

## Reversal triggers

This decision should be reconsidered if any of the following fire in Stage 1+:

1. **Slack publishes a vendor-maintained MCP server.** Migrate slack from `slack_sdk` to that server. Direct-SDK dispatch stays available as the fallback shape if the vendor server has gaps in the read-only enumeration surface Stage 0 needs.
2. **Google publishes a vendor-maintained MCP server** (for Drive specifically, or as part of a broader Google Workspace MCP). Same as Slack.
3. **A sixth source-kind needs adding** (beyond github/slack/gdrive plus whichever two land in Stage 1) and the per-source dispatch boilerplate compounds into a maintenance burden. At that point, evaluate either (a) generalizing the dispatch to a registry-driven plugin pattern or (b) collapsing to all-MCP if the ecosystem has matured (vendor MCP servers for each of the six).
4. **Docker becomes unavailable as a dev-stack prerequisite.** This could happen for offline-first installs, regulated enterprise environments with strict egress policies that block `ghcr.io` pulls, or a future shift in the project's prerequisite set. Migrate github from local Docker subprocess to the remote HTTP MCP server at `https://api.githubcopilot.com/mcp/` via the mcp Python SDK's HTTP transport, accepting the operational-posture trade-off that ships with vendor-hosted endpoints.

Until then, per-source dispatch handles Stage 0's Onboarding Crew.

## Consequences

**Operational.** One new Docker image pin to add to CLAUDE.md's service-image pins table: `ghcr.io/github/github-mcp-server:v1.0.4`. Node.js ≥20, which would have been required to host the deprecated `@modelcontextprotocol/server-github` npm package, is no longer a developer-machine prerequisite. Net -0 dev-stack prerequisites compared to the *original* plan (the original plan implicitly assumed Node was available); net simplification compared to where the inline plan landed before the discovery pass.

**Code.** Connector dispatches on `source_kind` to one of three code paths. Crawler walks each authenticated source through the same client the Connector instantiated. Credentials persist in a new Postgres table (`tenant_credentials`) under AES-256-GCM, with the key derived via HKDF-SHA256 from the Stage 0 Ed25519 dev keypair's private bytes. Standard primitives throughout; no novel cryptography. A Python Temporal worker on the `galileo-onboarding-crew` task queue is the new process introduced by PR-D.

**Plan & spec.** STAGE_0_PLAN.md §Week 4 deliverables 1 and 2 record the substitution. The gate-test requirement preserved unchanged (Onboarding Crew produces a valid manifest against the internal test workspace; manifest-check validator passes).

**Cross-pollination of the dependency-shape pattern.** This finding feeds back into [`docs/solutions/SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md`](../solutions/SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md) as instance 4 and contributes the **distribution-channel axis** to the pattern's decomposition (deployment topology / container topology / ecosystem utilities / distribution channel). Four axes, four named verification commands. The solutions doc compounds in structure, not just count.
