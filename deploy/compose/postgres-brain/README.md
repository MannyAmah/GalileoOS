# postgres-brain — custom Galileo Postgres image (pgvector + AGE)

First "thing Galileo builds itself" in the project. Maintenance posture is documented here so future contributors editing this Dockerfile see the policy alongside the code.

## What this image is

`apache/age:release_PG17_1.6.0` (Apache AGE on PostgreSQL 17, debian-bookworm base) **+** `postgresql-17-pgvector` 0.8.2 from PGDG.

Single image carrying both extensions Galileo's Brain needs (pgvector for semantic memory; AGE for relational/graph memory). Built in CI per [`CLAUDE.md`](../../../CLAUDE.md) "Service image pins"; not yet published to GHCR (reconsider when ≥3 custom images exist in `deploy/compose/`; currently 2).

## Pin policy

Two version pins live in this directory's Dockerfile:

| Pin | Current | Source of truth |
|---|---|---|
| Upstream base | `apache/age:release_PG17_1.6.0` | Docker Hub `apache/age` tag |
| pgvector apt | unpinned — latest from image's apt index at build time | PGDG repo configured by `postgres:17` base; resolved per build (PR-E first iteration: pinned to `0.8.2-1.pgdg12+1` failed because the apache/age image's apt index lagged the PGDG public snapshot read at planning time) |

Both pins co-change with the matching row in `CLAUDE.md` "Service image pins" in the same PR. Drift between the Dockerfile pin and the CLAUDE.md row is the same shape of mistake as the CI ↔ devcontainer co-change drift documented in CLAUDE.md "Tool version pins."

## Upstream-base-version-bump procedure

When `apache/age` publishes a newer Docker Hub tag (e.g., `release_PG17_1.7.0` when it Dockerizes from the existing GitHub tag), or PGDG publishes a newer pgvector point release:

1. **Open a dedicated PR** (do not bundle with feature work).
2. Update the `FROM` line and/or the `apt-get install` pin in this Dockerfile.
3. Update the matching row in `CLAUDE.md` "Service image pins."
4. Run `make up` locally and verify the Brain integration tests pass (`go test -tags=gateway_integration -count=1 -v ./cmd/gateway/...`).
5. Surface the upstream changelog notes in the PR description (what changed in the new release; whether any breaking changes affect Galileo's usage of either extension).
6. If the new AGE major adds breaking schema changes, evaluate a separate migration to handle the transition; do not silently absorb.

## Reconsideration triggers

- **GHCR publishing.** Build-in-CI is the current pattern (PR-E precedent). Promote to GHCR publishing when ≥3 custom images exist in `deploy/compose/`. Currently 2 (`otel-wrapper/` from PR-B, `postgres-brain/` from PR-E).
- **AGE 1.7.0 Dockerization.** A GitHub tag `PG17/v1.7.0-rc0` exists from 2026-02-11 but is not yet on Docker Hub. When the `apache/age:release_PG17_1.7.0` Docker Hub tag publishes, evaluate upgrade in a separate PR per the procedure above.
- **Debian distro shift.** apache/age inherits from `postgres:17`, which is debian-bookworm. If the upstream `postgres:17` floating tag moves to debian-trixie, the PGDG version string changes (`pgdg12+1` → `pgdg13+1`). Re-verify the pgvector pin during the bump.

## What this directory is *not*

- Not a publication target. The image lives in CI scratch space and developer Docker daemons; no `docker push` happens.
- Not a long-running container source. The compose service uses this image as its base, but the image itself doesn't define startup behavior — that's `apache/age`'s base entrypoint.
- Not Galileo's official Brain server. The Brain is a *substrate* (the extensions running inside Postgres), not a separate service. Galileo agents query the same Postgres that hosts `tenants` and `cost_events`.
