# SOLUTION — Dependency shape verification (the marketing description is not the deployment shape)

Three findings from three separate inline-plan rounds, each one a case where the *shape* of a dependency — its deployment topology, its packaging, what's actually in its image — differed from how the plan named it. The first two were caught at planning time and prevented bad implementation; the third was missed at planning time and surfaced as two rework CI iterations during PR-A. Three instances is enough to make this a documented pattern rather than scattered observations.

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

## Generalization

**The marketing description of a dependency is not its deployment shape.** "Observability platform," "data-plane substrate," "LLM proxy image" — these are product labels. The deployment shape is what `docker pull && docker run --entrypoint sh <image> -c 'ls /usr/bin'` returns, what `wc -l docker-compose.yml` in the dependency's self-hosting docs reads, what the deployment-model section of the dependency's README actually says about embedding vs. server vs. sidecar.

**Shape verification at planning time is cheap.** A `gh api` call, an image pull and listing, a careful read of the dependency's "how to deploy this" section. Each finding above took less than ten minutes of planning-time work that would have caught the issue. The cost shows up later: a plan-deviation PR rewriting an entire layer's placement (instance 1), a deferred substrate decision (instance 2), two CI rework iterations during what should have been a clean review-and-merge (instance 3).

**Where to apply this discipline.** Before any inline plan that names an external service or container image:

1. **For libraries that ship as SDKs:** read the deployment-model section of the project's README. Confirm language coverage (SDK, native server, sidecar). The Mirage SDK-only finding came from one paragraph.
2. **For services that ship as deployments:** count the containers in the self-hosting docker-compose. If the count is one or two, the plan's framing of "a service" is probably accurate. If it's eight or thirteen, the framing was wrong and the scope is different.
3. **For container images we depend on:** fetch the upstream Dockerfile (`gh api repos/<owner>/<repo>/contents/Dockerfile` decodes from base64) and confirm what's installed in the *runtime* stage. Especially relevant for minimal-base images (Chainguard wolfi-base, distroless, scratch) where standard utilities may be absent.

**What doesn't earn this discipline.** Internal services, libraries already vendored, dependencies named in this CLAUDE.md. The verification is for *new* external dependencies entering the plan — that's where the marketing description is the only thing the planner has seen.

## Related lessons

- [SOLUTION — CI guard verification](SOLUTION_CI_GUARD_VERIFICATION.md) — workflow-file defects propagate across all concurrent PRs (different failure mode, same family: assumptions about external machinery that need verification before the assumption is committed to).
- [SOLUTION — Dependency coupling](SOLUTION_DEPENDENCY_COUPLING.md) — once a dependency is adopted, version coupling across consumers compounds; shape verification at planning time keeps the dependency graph honest before that coupling sets in.
