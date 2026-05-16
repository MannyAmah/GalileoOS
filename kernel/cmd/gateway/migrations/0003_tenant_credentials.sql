-- 0003_tenant_credentials: encrypted per-source credentials for the
-- Onboarding Crew (PR-D / Week 4).
--
-- One row per (tenant_id, source_kind). encrypted_payload is the full
-- AES-256-GCM ciphertext written by agents.onboarding.credentials —
-- the 12-byte GCM nonce is prepended to the ciphertext, so the column
-- is self-contained for decryption; no separate nonce column needed.
--
-- The AES key derives via HKDF-SHA256 from the Stage 0 Ed25519 dev
-- keypair's private bytes. The associated data bound into the GCM tag
-- is "{tenant_id}:{source_kind}", so a row lifted to a different
-- (tenant, source) pair fails to decrypt — defense in depth against
-- row-level credential confusion.
--
-- updated_at supports the CLI's ON CONFLICT DO UPDATE path so
-- operators can re-run with rotated credentials and see when the
-- last rotation landed.

CREATE TABLE IF NOT EXISTS tenant_credentials (
    tenant_id          UUID        NOT NULL REFERENCES tenants(tenant_id),
    source_kind        TEXT        NOT NULL CHECK (source_kind IN ('github', 'slack', 'gdrive')),
    encrypted_payload  BYTEA       NOT NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, source_kind)
);
