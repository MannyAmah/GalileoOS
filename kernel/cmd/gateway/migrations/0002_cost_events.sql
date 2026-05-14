-- 0002_cost_events: per-LLM-call cost ingestion table.
--
-- Written by the gateway's POST /internal/cost-events webhook handler
-- when LiteLLM's generic_api callback POSTs a StandardLoggingPayload.
-- Read by the budget-cap middleware (sum cost_cents per tenant per
-- current month) and by the cost-recon binary that reconciles against
-- Stripe metered billing.
--
-- request_id is the PRIMARY KEY because the webhook needs idempotency:
-- LiteLLM may re-deliver the same callback, recon may re-run, and the
-- gateway must never double-count. request_id is the gateway-generated
-- galileo_request_id (UUIDv7) injected into LiteLLM's request body
-- metadata; the webhook reads it back from
-- payload.metadata.requester_metadata.galileo_request_id.
--
-- litellm_id is informational only — LiteLLM's own call id from
-- payload.id, kept for upstream debugging but not used for joins.

CREATE TABLE IF NOT EXISTS cost_events (
    request_id    TEXT        PRIMARY KEY,
    tenant_id     UUID        NOT NULL REFERENCES tenants(tenant_id),
    event_ts      TIMESTAMPTZ NOT NULL,
    cost_cents    BIGINT      NOT NULL CHECK (cost_cents >= 0),
    provider      TEXT,
    model         TEXT,
    litellm_id    TEXT
);

-- Budget-cap middleware reads sum(cost_cents) for the current month per
-- tenant on every request — index makes that read O(matching rows)
-- instead of O(table) once tenants accumulate history.
CREATE INDEX IF NOT EXISTS cost_events_tenant_ts_idx
    ON cost_events (tenant_id, event_ts DESC);
