# Stage 1 Brain substrate (PR-E / ADR-0006). pgvector + Apache AGE on
# PostgreSQL 17 in one image. Base: apache/age:release_PG17_1.6.0 (which
# inherits FROM postgres:17 / debian-bookworm and ships AGE preinstalled
# with shared_preload_libraries='age' preconfigured). Adds pgvector from
# the PGDG apt repo (already configured in the postgres:17 base via the
# inherited apt sources.list.d entries).
#
# Pin policy: this Dockerfile and the CLAUDE.md "Service image pins" row
# for `postgres-brain` co-change in the same PR. Bumping the upstream
# apache/age base or the pgvector apt version is a deliberate PR-level
# change, not a rebuild side effect. See README.md in this directory
# for the upstream-base-version-bump procedure.
#
# Runtime-introduction size exception, 2nd instance (PR-A was the 1st).
# Cited in PR-E's description per the "Size exception discipline"
# section in CLAUDE.md.

FROM apache/age:release_PG17_1.6.0

# pgvector 0.8.2 from PGDG. Explicit version pin so CI rebuilds are
# reproducible — debian-bookworm's PGDG index updates do not silently
# shift the version installed. Verified 2026-05-16 via
# https://apt.postgresql.org/pub/repos/apt/dists/bookworm-pgdg/main/binary-amd64/Packages.gz
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
       postgresql-17-pgvector=0.8.2-1.pgdg12+1 \
    && rm -rf /var/lib/apt/lists/*
