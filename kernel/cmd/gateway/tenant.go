// Tenant resolution for the Stage 0 gateway.
//
// Drift-1 (Week 3 inline-plan resolution): monthly_budget_cents is read
// fresh from Postgres on every request. The JWT carries an informational
// copy of the value, but the authoritative number is whatever Postgres
// returns *right now*. If Postgres is unreachable the request is denied
// with 503 — we never serve on the stale JWT-cached value.
//
// The kernel/auth package handles JWT verification; this file handles
// the Postgres lookup that runs after verification succeeds.

package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/MannyAmah/GalileoOS/kernel/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrPostgresUnavailable signals the tenant store is unreachable. The
// gateway's auth middleware translates this to HTTP 503; any other error
// from Resolve is treated as a token-level failure (401).
var ErrPostgresUnavailable = errors.New("postgres unavailable")

// TenantContext carries the resolved tenant identity plus the
// Postgres-fresh budget value used by downstream handlers (and by PR-B's
// budget-cap check).
type TenantContext struct {
	TenantID           string
	MonthlyBudgetCents int64
}

// TenantResolver verifies a bearer token and looks up the tenant's
// current budget from Postgres. It is safe for concurrent use; the pool
// handles connection lifecycle.
type TenantResolver struct {
	pool *pgxpool.Pool
}

// NewTenantResolver wraps an existing pgx pool. Callers own the pool's
// lifecycle (Close is not called by the resolver).
func NewTenantResolver(pool *pgxpool.Pool) *TenantResolver {
	return &TenantResolver{pool: pool}
}

// Resolve verifies the JWT against the public key at pubKeyPath, then
// reads monthly_budget_cents from Postgres for that tenant. The returned
// context carries the TenantContext for downstream handlers.
//
// Drift-1: if Postgres is unreachable, returns ErrPostgresUnavailable.
// The caller (auth middleware) maps that to 503. A missing tenant row
// is *not* ErrPostgresUnavailable — it's a 401 (the JWT pointed at a
// tenant that does not exist).
func (r *TenantResolver) Resolve(ctx context.Context, pubKeyPath, raw string) (context.Context, error) {
	claims, err := auth.VerifyToken(pubKeyPath, raw)
	if err != nil {
		return nil, fmt.Errorf("verify token: %w", err)
	}

	var budget int64
	row := r.pool.QueryRow(ctx,
		`SELECT monthly_budget_cents FROM tenants WHERE tenant_id = $1`,
		claims.TenantID,
	)
	if err := row.Scan(&budget); err != nil {
		// Row not found is a token-level failure: the JWT pointed at a
		// tenant that does not exist. Returned as 401 by the caller.
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("tenant %s not found", claims.TenantID)
		}
		// Any other error means we could not authoritatively read the
		// budget — pool closed, connection refused, query timeout, etc.
		// Drift-1: we never serve on stale data, so this is 503.
		return nil, ErrPostgresUnavailable
	}

	tc := TenantContext{
		TenantID:           claims.TenantID,
		MonthlyBudgetCents: budget,
	}
	return context.WithValue(ctx, tenantCtxKey{}, tc), nil
}

type tenantCtxKey struct{}

// TenantFromContext returns the TenantContext attached by Resolve. Panics
// if no tenant is on the context — handlers behind authMiddleware are
// guaranteed one, and a missing tenant is a programmer error, not a
// runtime condition.
func TenantFromContext(ctx context.Context) TenantContext {
	tc, ok := ctx.Value(tenantCtxKey{}).(TenantContext)
	if !ok {
		panic("gateway: TenantFromContext called outside authMiddleware")
	}
	return tc
}
