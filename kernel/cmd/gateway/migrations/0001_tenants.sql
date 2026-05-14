-- 0001_tenants: tenant table introduced in PR #15 / PR-A.
--
-- PR-A's integration test created this table inline via CREATE TABLE IF
-- NOT EXISTS. PR-B promotes that ad-hoc DDL to a versioned migration so
-- the gateway boot path is the canonical place schema lives, and so
-- production rollouts get the same DDL as the integration test.

CREATE TABLE IF NOT EXISTS tenants (
    tenant_id UUID PRIMARY KEY,
    monthly_budget_cents BIGINT NOT NULL CHECK (monthly_budget_cents >= 0)
);
