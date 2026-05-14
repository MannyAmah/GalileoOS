# `deploy/compose/` — Stage 0 single-VM docker-compose stack

`docker-compose.yml` brings up the runtime dependencies needed by `galileo-gateway` (and, in later Week 3 PRs, the agent-runner and web UI).

## What ships when

| PR | Adds to compose | Removes from compose |
|---|---|---|
| PR-A (this) | `postgres:17.9-alpine`, `temporalio/auto-setup:1.29.6.1`, `ghcr.io/berriai/litellm:v1.83.14-stable.patch.3` | — |
| PR-B | Observability backend (substrate decided in PR-B's planning round per `docs/plans/STAGE_0_PLAN.md` §Week 3 deliverable 7) | — |
| PR-C | `galileo-agent-runner`, `galileo-web` (as proper services rather than local processes) | — |

## What's deliberately NOT here

- **Opik.** The original plan named Opik for observability. The Week 3 inline-plan discovery pass found Opik's self-hosted deployment is a 13-container platform (MySQL + Redis + ClickHouse + Zookeeper + MinIO + 5 Opik-specific services + Jaeger + OTel collector). Substrate decision deferred to PR-B; if a lighter substitution is chosen there, it lands as a plan-deviation per the four-part template established in PR #13.
- **`galileo-gateway` as a containerized service.** PR-A runs the gateway as a local subprocess (or as the test binary in CI) against the services above. Containerizing the gateway is part of the Week 4 Stage 0 gate-test path.

## Pin policy

Image tags are pinned per CLAUDE.md "Service image pins"; co-changed when bumped under the same discipline as toolchain pins. Service-image pins do **not** follow the latest-1 language posture — they track latest stable directly, because they are runtime infrastructure not language baselines.

## Running

```bash
make up                # bring up the stack (compose v2 syntax)
docker compose -f deploy/compose/docker-compose.yml ps      # service status
docker compose -f deploy/compose/docker-compose.yml down -v # tear down + clear postgres volume
```
