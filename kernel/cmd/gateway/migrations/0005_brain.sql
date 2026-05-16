-- 0005_brain: Stage 1 Brain substrate (PR-E / ADR-0006).
--
-- Three kinds of memory, one substrate (per docs/galileo_os_infrastructure_plan.md §4.4):
--   - Semantic memory  → pgvector embeddings
--   - Episodic memory  → append-only events table
--   - Relational/graph → Apache AGE graph
--
-- All three ship in one migration because the substrate is the substrate;
-- downstream agents (Ingestion, Org-Mapper, Skill-Selector, QA) consume
-- different slices but shouldn't carry schema migrations as part of their
-- feature work. Splitting the substrate across feature PRs creates
-- artificial dependencies between agents that don't otherwise interact.
--
-- Extension requirements:
--   - vector  — pgvector, installed in postgres-brain image (PR-E)
--   - age     — Apache AGE, installed in postgres-brain image (PR-E)
--
-- AGE setup is multi-step: LOAD, search_path, create_graph. The
-- ALTER DATABASE sets the search_path for every future connection so
-- Brain-aware Go and Python code doesn't need per-connection setup.
-- AGE catalog writes from create_graph commit with this transaction;
-- subsequent connections see brain_graph immediately post-commit.

CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS age;
LOAD 'age';

-- Make ag_catalog available on every future connection without per-
-- connection SET search_path discipline. Safe because ag_catalog
-- doesn't shadow names in public or user schemas under normal usage.
ALTER DATABASE galileo SET search_path = ag_catalog, "$user", public;

-- Schema-qualify Galileo-owned tables per the CLAUDE.md
-- "Schema-qualification convention (Stage 1+)" section. ag_catalog is
-- first in search_path post-ALTER DATABASE above; explicit `public.`
-- prefix guards against future ag_catalog name collisions.

-- ----------------------------------------------------------------------
-- Semantic memory: per-chunk embeddings with metadata
-- ----------------------------------------------------------------------
-- Embedding dimensionality matches bge-large-en-v1.5 (1024 dim) per
-- canonical plan §4.4 ingestion pipeline. If the embedding model
-- changes in Stage 2, a separate migration adds a new column or
-- replaces this table; the dimensionality is part of the schema
-- contract.
CREATE TABLE IF NOT EXISTS public.brain_embeddings (
    embedding_id    UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID         NOT NULL REFERENCES public.tenants(tenant_id),
    source_kind     TEXT         NOT NULL,
    source_uri      TEXT         NOT NULL,
    chunk_text      TEXT         NOT NULL,
    embedding       vector(1024) NOT NULL,
    metadata        JSONB,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS brain_embeddings_tenant_idx
    ON public.brain_embeddings (tenant_id);

-- ivfflat is the default index type for pgvector cosine similarity
-- on tables under ~1M rows. Switch to hnsw when a tenant exceeds the
-- ivfflat sweet spot per Stage 2 trigger (deferral table row would
-- be added at that time).
CREATE INDEX IF NOT EXISTS brain_embeddings_vector_idx
    ON public.brain_embeddings USING ivfflat (embedding vector_cosine_ops);

-- ----------------------------------------------------------------------
-- Episodic memory: time-stamped append-only events scoped to tenant
-- ----------------------------------------------------------------------
-- Reads: "Last Tuesday Jane in Support escalated a billing issue from
-- ACME and resolved it by issuing a credit." Stored as structured rows
-- so the Org-Mapper Agent can synthesize timeline-aware summaries.
CREATE TABLE IF NOT EXISTS public.brain_events (
    event_id        UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID         NOT NULL REFERENCES public.tenants(tenant_id),
    event_ts        TIMESTAMPTZ  NOT NULL,
    actor           TEXT,
    action          TEXT         NOT NULL,
    subject         TEXT,
    payload         JSONB,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS brain_events_tenant_ts_idx
    ON public.brain_events (tenant_id, event_ts DESC);

-- ----------------------------------------------------------------------
-- Relational/graph memory: AGE graph for entity relationships
-- ----------------------------------------------------------------------
-- Org-Mapper Agent's primary substrate. Cypher queries against this
-- graph synthesize org charts, customer relationships, vendor
-- networks. Per-tenant graph isolation lives at the application
-- layer for Stage 1 (tenant_id label on every vertex); promoted to
-- per-tenant graphs if scale requires (Stage 2+ trigger).
--
-- Fully qualified `ag_catalog.create_graph` (not bare `create_graph`)
-- because pgx's single-Exec multi-statement protocol doesn't activate
-- the search_path change above until the transaction commits — within
-- the same tx the function-name resolution falls back to the original
-- search_path. Same shape as the AGE README's autocommit-disabled
-- client warning. PR-E iteration 2 finding: the bare call returns
-- SQLSTATE 42883 (function does not exist) inside the migration tx.
SELECT ag_catalog.create_graph('brain_graph');
