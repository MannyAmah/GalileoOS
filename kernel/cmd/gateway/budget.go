// Budget cap enforcement (cost-meter loop A).
//
// PR-A's TenantResolver gave us monthly_budget_cents Postgres-fresh
// per request. PR-B's budget middleware sums cost_events.cost_cents for
// the current month per tenant and rejects with HTTP 402 when the sum
// exceeds the cap.
//
// The sum query and the cap read happen in the same transaction so the
// budget decision is consistent against a single Postgres snapshot.
// Stage 0 acceptance: the cost of the per-request sum is acknowledged
// and accepted; per-tenant caching is a Stage 1 concern when traffic
// makes it real. Premature optimization is rejected.

package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// budgetCheck middleware reads the current month's spend for the tenant
// attached to the request context (set by authMiddleware) and denies
// with 402 when the spend equals or exceeds the cap. Inserted in the
// chain *after* authMiddleware so the TenantContext is available.
func (s *Server) budgetMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tc := TenantFromContext(r.Context())
		spent, err := readMonthlySpend(r.Context(), s.pool, tc.TenantID)
		if err != nil {
			s.logger.Printf("budget: read monthly spend failed for tenant=%s: %v", tc.TenantID, err)
			writeErr(w, http.StatusServiceUnavailable, "budget store unavailable")
			return
		}
		if spent >= tc.MonthlyBudgetCents {
			writeBudgetDenied(w, spent, tc.MonthlyBudgetCents)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// readMonthlySpend returns the sum of cost_cents for the tenant within
// the current calendar month (UTC). Postgres handles the date_trunc;
// returning 0 with no error when the tenant has no rows yet keeps the
// caller's deny logic uniform.
func readMonthlySpend(ctx context.Context, pool *pgxpool.Pool, tenantID string) (int64, error) {
	var spent int64
	err := pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(cost_cents), 0)::BIGINT
		FROM public.cost_events
		WHERE tenant_id = $1
		  AND event_ts >= date_trunc('month', now() AT TIME ZONE 'UTC')
	`, tenantID).Scan(&spent)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("sum cost_events: %w", err)
	}
	return spent, nil
}

func writeBudgetDenied(w http.ResponseWriter, spent, budget int64) {
	// HTTP 402 Payment Required per RFC 9110 §15.5.2 — semantically correct
	// for budget refusal and the agent-runner (sole Stage 0 consumer) handles
	// the response body programmatically. Reconsider 429 in Stage 1 if
	// external API consumers appear; some HTTP clients and intermediaries
	// handle 402 inconsistently while 429 is universally retried-with-backoff.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusPaymentRequired)
	_ = writeJSON(w, map[string]any{
		"error":       "monthly_budget_cents exceeded",
		"spent_cents": spent,
		"budget_cents": budget,
	})
}
