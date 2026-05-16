-- 0004_tenant_manifests: per-source manifest rows emitted by the
-- Crawler Agent (PR-D / Week 4).
--
-- One row per (tenant_id, source_kind). crawl_status moves
-- 'in_progress' -> 'completed' | 'failed'. manifest_json holds the
-- enumerated per-document records (path, size_bytes, content_type,
-- last_modified_unix) — a JSON array; Stage 1's Ingestion Agent reads
-- this column as its input contract.
--
-- document_count and total_bytes are materialized aggregates so the
-- manifest-check Go binary can validate gate dimensions (50K-docs
-- cap, expected vs actual counts) without parsing JSON.
--
-- manifest_id is a separate PK so future iterations can keep history
-- (one row per crawl run) — Stage 0 upserts on (tenant_id,
-- source_kind), so the unique index below enforces one row per pair.

CREATE TABLE IF NOT EXISTS tenant_manifests (
    manifest_id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            UUID         NOT NULL REFERENCES tenants(tenant_id),
    source_kind          TEXT         NOT NULL CHECK (source_kind IN ('github', 'slack', 'gdrive')),
    crawl_status         TEXT         NOT NULL DEFAULT 'in_progress'
                                      CHECK (crawl_status IN ('in_progress', 'completed', 'failed')),
    crawl_started_at     TIMESTAMPTZ,
    crawl_completed_at   TIMESTAMPTZ,
    document_count       BIGINT       NOT NULL DEFAULT 0,
    total_bytes          BIGINT       NOT NULL DEFAULT 0,
    manifest_json        JSONB
);

-- Stage 0 upsert pattern is ON CONFLICT (tenant_id, source_kind);
-- this unique index enforces the one-row-per-pair invariant the
-- Crawler relies on.
CREATE UNIQUE INDEX IF NOT EXISTS tenant_manifests_tenant_source_uidx
    ON tenant_manifests (tenant_id, source_kind);
