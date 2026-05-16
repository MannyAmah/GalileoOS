# SOLUTION — Dependency shape verification (the marketing description is not the deployment shape)

Four findings from four separate inline-plan rounds, each one a case where the *shape* of a dependency — its deployment topology, its packaging, what's actually in its image, how it's distributed — differed from how the plan named it. The first two were caught at planning time and prevented bad implementation; the third was missed at planning time and surfaced as two rework CI iterations during PR-A; the fourth was caught by the discovery pass at planning time for PR-D and contributed a previously-unnamed axis of the same pattern. Four instances is enough to decompose the pattern by axis and name a verification command per axis.

## Instance 1 — Mirage's deployment model (PR #13, 2026-05-13)

**Plan framing.** Mirage placed at Galileo's Layer 3 (Go kernel) as the unified data-plane substrate, with a Stage 0 live probe gating adoption.

**Actual shape, surfaced by reading Mirage's docs directly.** "[Mirage] ships Python and TypeScript SDKs only, with no Go SDK and no native server. Mirage is designed for in-process embedding inside agent code (FastAPI, Express, browser apps, async runtimes)." Placing it at the Go kernel layer would have required a permanent Python sidecar not named in the original plan.

**Outcome.** First plan-deviation in the project's history. Mirage relocated to Layer 5 (integrations); the live probe was no longer the right experiment because the structural mismatch was already decisive. See [`docs/decisions/0003-mirage-layer-relocation.md`](../decisions/0003-mirage-layer-relocation.md) and [`docs/closeouts/CLOSEOUT_LAYER3_MIRAGE_RECONSIDERED.md`](../closeouts/CLOSEOUT_LAYER3_MIRAGE_RECONSIDERED.md).

## Instance 2 — Opik's container count (PR #15 / Week 3 inline plan, 2026-05-14)

**Plan framing.** Opik named in the canonical plan as the observability backend at Layer 5, deliverable 7. Implicit framing: comparable in weight to Postgres or LiteLLM (a single service container).

**Actual shape, surfaced by reading Opik's self-hosted deployment documentation.** "Opik's self-hosted deployment is a 13-container platform (MySQL + Redis + ClickHouse + Zookeeper + MinIO + 5 Opik-specific services + Jaeger + OTel collector)." Not a single observability service.

**Outcome.** Substrate decision deferred to PR-B's planning round, recorded in [`docs/plans/STAGE_0_PLAN.md`](../plans/STAGE_0_PLAN.md) §Week 3 deliverable 7. PR-A explicitly does *not* pre-commit a substitution — that's PR-B's plan-deviation work under the four-part template (closeout + plan edits + ADR + code).

## Instance 3 — wolfi-base utility absence (PR #15 CI, 2026-05-14)

**Plan framing.** LiteLLM v1.83.14-stable.patch.3 pinned in compose YAML and CI YAML; healthcheck used `wget -qO- http://localhost:4000/health/liveliness`. Implicit assumption: standard utilities are present in the runtime image.

**Actual shape, surfaced by reading LiteLLM's Dockerfile.** The runtime stage is built on `cgr.dev/chainguard/wolfi-base` and installs only `bash, openssl, tzdata, nodejs, npm, python3, libsndfile` — **neither curl nor wget**. The first CI iteration's `wget` failure and the second iteration's `curl` retry both surfaced the same root cause: the binary the healthcheck named was not in the image. Took two CI iterations during PR-A to surface what `gh api repos/BerriAI/litellm/contents/Dockerfile` answered in one call at planning time.

**Outcome.** Healthcheck switched to `bash -c '</dev/tcp/127.0.0.1/4000'` in both compose YAML and CI YAML — bash IS in the wolfi runtime, and Uvicorn binds the port only after `Application startup complete` per the boot logs, so port-open is a meaningful ready signal.

## Instance 4 — `github-mcp-server` distribution channel (PR-D / Week 4 inline plan, 2026-05-16)

**Plan framing.** The Week 4 plan named `@modelcontextprotocol/server-github` invoked via `npx -y` as one of three reference MCP servers for the Onboarding Crew. Implicit framing: an npm-distributed Node.js subprocess, available via the standard `npx` invocation pattern, parallel to the slack and gdrive reference servers.

**Actual shape, surfaced by reading GitHub's vendor-maintained MCP server release artifacts and installation docs.** Two compounding findings: (a) the three `@modelcontextprotocol/server-*` reference packages were archived upstream on commit `d53d6cc75c` of `modelcontextprotocol/servers` on 2025-05-29; (b) GitHub's vendor-maintained replacement at `github/github-mcp-server` (v1.0.4, 29.8k stars) is distributed as a Docker image (`ghcr.io/github/github-mcp-server`) and Go binary release tarballs — **not as an npm package.** The release-asset list contains `github-mcp-server_Darwin_arm64.tar.gz`, `github-mcp-server_Linux_x86_64.tar.gz`, and equivalents — zero npm artifacts. GitHub's own installation docs lead with the Docker invocation pattern; npm is simply not the channel.

**Outcome.** Third plan-deviation in the project's history. Per-source dispatch: Docker-subprocess MCP for github (`docker run -i --rm --init -e GITHUB_PERSONAL_ACCESS_TOKEN ghcr.io/github/github-mcp-server:v1.0.4`); direct SDK (`slack_sdk`, `google-api-python-client`) for slack and gdrive. See [`docs/decisions/0005-mcp-per-source-vs-mixed.md`](../decisions/0005-mcp-per-source-vs-mixed.md) and [`docs/closeouts/CLOSEOUT_MCP_PER_SOURCE_RECONSIDERED.md`](../closeouts/CLOSEOUT_MCP_PER_SOURCE_RECONSIDERED.md).

## Generalization

**The marketing description of a dependency is not its deployment shape.** "Observability platform," "data-plane substrate," "LLM proxy image," "MCP server" — these are product labels. The deployment shape is what `docker pull && docker run --entrypoint sh <image> -c 'ls /usr/bin'` returns, what `wc -l docker-compose.yml` in the dependency's self-hosting docs reads, what the deployment-model section of the dependency's README actually says about embedding vs. server vs. sidecar, and what registry the release artifacts actually publish to.

**Four axes of the same underlying pattern.** The four documented instances are not interchangeable — each surfaces along a different axis. Future planning rounds reach for the right verification command per concern:

| Axis | Instance | Verification command at planning time |
|---|---|---|
| **Deployment topology** (SDK vs server vs sidecar) | 1 — Mirage SDK-only | Read the deployment-model section of the project's README. Confirm language coverage. |
| **Container topology** (one service vs many) | 2 — Opik 13-container platform | `wc -l docker-compose.yml` in the project's self-hosting repo, or count `services:` entries with `grep -c '^[[:space:]][[:space:]][a-z].*:$'`. |
| **Ecosystem utilities** (what's actually in the runtime image) | 3 — wolfi-base missing curl/wget | `docker run --rm --entrypoint sh <image> -c 'ls /usr/bin'`, or fetch the upstream Dockerfile via `gh api repos/<owner>/<repo>/contents/Dockerfile` (base64-decoded). |
| **Distribution channel** (npm vs Docker vs Go binary vs PyPI vs hosted endpoint) | 4 — github-mcp-server not on npm | Query the assumed registry: `npm view <name>`, `pip show <name>` or `curl -sL https://pypi.org/pypi/<name>/json`, `gh release list -R <owner>/<repo>`, `docker manifest inspect <image>`. If the assumed channel returns 404 or "deprecated," the assumption was wrong. |

**Shape verification at planning time is cheap.** A `gh api` call, an image pull and listing, a registry query, a careful read of the dependency's "how to deploy this" section. Each finding above took less than ten minutes of planning-time work that would have caught the issue. The cost shows up later: a plan-deviation PR rewriting an entire layer's placement (instance 1), a deferred substrate decision (instance 2), two CI rework iterations during what should have been a clean review-and-merge (instance 3), a third plan-deviation rewriting the connector dispatch model (instance 4).

**Where to apply this discipline.** Before any inline plan that names an external service, container image, or registry-distributed package, run the appropriate verification command from the axis table above. The axis is selected by what the plan is asserting about the dependency: if the plan asserts "we'll embed this library," verify deployment topology; if it asserts "we'll run this service," verify container topology; if it asserts "we'll healthcheck this container with `wget`," verify ecosystem utilities; if it asserts "we'll install this via `npx`/`pip`/`brew`," verify distribution channel.

**What doesn't earn this discipline.** Internal services, libraries already vendored, dependencies named in this CLAUDE.md. The verification is for *new* external dependencies entering the plan — that's where the marketing description is the only thing the planner has seen.

## Related lessons

- [SOLUTION — CI guard verification](SOLUTION_CI_GUARD_VERIFICATION.md) — workflow-file defects propagate across all concurrent PRs (different failure mode, same family: assumptions about external machinery that need verification before the assumption is committed to).
- [SOLUTION — Dependency coupling](SOLUTION_DEPENDENCY_COUPLING.md) — once a dependency is adopted, version coupling across consumers compounds; shape verification at planning time keeps the dependency graph honest before that coupling sets in.
