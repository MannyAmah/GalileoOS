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

# pgvector via apt. Pin posture for PR-E: install latest from the
# apt index available inside the apache/age:release_PG17_1.6.0 image.
# An exact-version pin (e.g., `0.8.2-1.pgdg12+1`) was tried first and
# rejected because the version available in this image's apt index
# can lag or lead the PGDG bookworm-pgdg snapshot we read at
# planning time. The looser-pin posture trades reproducibility for
# build resilience.
#
# When the GHCR-publishing trigger fires (CLAUDE.md "Custom Galileo-
# built images" → ≥3 custom images), we shift to a build-once-then-
# tag posture where the exact installed version is captured in the
# published tag.
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
       postgresql-17-pgvector \
    && rm -rf /var/lib/apt/lists/*
