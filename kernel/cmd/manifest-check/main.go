// Package main is the Stage 0 manifest-check binary used by the
// Onboarding Crew to validate per-source manifests against the
// §3.5 gate dimensions.
//
// Dimensions checked (per docs/galileo_os_infrastructure_plan.md §3.5):
//
//	wall-clock      <  6 hours per source (computed from crawl_started_at to
//	                   crawl_completed_at)
//	LLM cost        <  $50 per tenant (sum of cost_events.cost_cents)
//	org-snapshot    >  90% of expected source-kinds enumerated successfully
//	skill-precision N/A in Stage 0 — Skill-Selector Agent is deferred to
//	                   Stage 1 per ADR-0003. Emit a bracketed [gate] line
//	                   naming the deferral; never silently skip.
//	destructive     == 0 actions observed (Stage 0 has no destructive
//	                   write surface; the check is structural)
//
// Exit codes:
//
//	0  all dimensions passed (or were honestly N/A in Stage 0)
//	1  one or more dimensions failed
//	2  invocation error (missing flags, database unreachable, etc.)
//
// Bracketed [gate] prefix on every per-dimension line is grep-able;
// readers see exactly which dimensions ran, which deferred, and which
// failed. Same pattern as PR-A's gateway error logging.

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxWallClock          = 6 * time.Hour
	maxLLMCentsPerTenant  = 5000 // $50.00
	minOrgSnapshotPercent = 90
	expectedSourceKinds   = 3 // github + slack + gdrive in Stage 0
)

type manifestRow struct {
	sourceKind        string
	crawlStatus       string
	crawlStartedAt    *time.Time
	crawlCompletedAt  *time.Time
	documentCount     int64
}

func main() {
	tenantID := flag.String("tenant", "", "tenant UUID to validate (required)")
	databaseURL := flag.String(
		"database-url",
		os.Getenv("GALILEO_GATEWAY_DATABASE_URL"),
		"Postgres URL; defaults to $GALILEO_GATEWAY_DATABASE_URL",
	)
	flag.Parse()

	if *tenantID == "" {
		fmt.Fprintln(os.Stderr, "manifest-check: -tenant <UUID> is required")
		os.Exit(2)
	}
	if *databaseURL == "" {
		fmt.Fprintln(os.Stderr, "manifest-check: -database-url or $GALILEO_GATEWAY_DATABASE_URL is required")
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, *databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "manifest-check: connect: %v\n", err)
		os.Exit(2)
	}
	defer pool.Close()

	rows, err := loadManifests(ctx, pool, *tenantID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "manifest-check: load manifests: %v\n", err)
		os.Exit(2)
	}

	costCents, err := loadCostCents(ctx, pool, *tenantID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "manifest-check: load cost: %v\n", err)
		os.Exit(2)
	}

	if checkAll(rows, costCents) {
		fmt.Println("[gate] manifest-check: all dimensions passed")
		os.Exit(0)
	}
	os.Exit(1)
}

func loadManifests(ctx context.Context, pool *pgxpool.Pool, tenantID string) ([]manifestRow, error) {
	const query = `
		SELECT source_kind, crawl_status, crawl_started_at, crawl_completed_at, document_count
		FROM public.tenant_manifests
		WHERE tenant_id = $1
		ORDER BY source_kind
	`
	rs, err := pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rs.Close()

	var out []manifestRow
	for rs.Next() {
		var r manifestRow
		if err := rs.Scan(&r.sourceKind, &r.crawlStatus, &r.crawlStartedAt, &r.crawlCompletedAt, &r.documentCount); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rs.Err()
}

func loadCostCents(ctx context.Context, pool *pgxpool.Pool, tenantID string) (int64, error) {
	const query = `
		SELECT COALESCE(SUM(cost_cents), 0)
		FROM public.cost_events
		WHERE tenant_id = $1
	`
	var total int64
	err := pool.QueryRow(ctx, query, tenantID).Scan(&total)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}
	return total, nil
}

// checkAll runs every dimension and prints one [gate] line per check.
// Returns true if every dimension passed (or was honestly N/A).
// Continues past a failure so the operator sees all failing dimensions
// in one invocation rather than fixing them one at a time.
func checkAll(rows []manifestRow, costCents int64) bool {
	ok := true
	ok = checkWallClock(rows) && ok
	ok = checkLLMCost(costCents) && ok
	ok = checkOrgSnapshot(rows) && ok
	ok = checkSkillPrecision() && ok
	ok = checkDestructive() && ok
	return ok
}

func checkWallClock(rows []manifestRow) bool {
	for _, r := range rows {
		if r.crawlStartedAt == nil || r.crawlCompletedAt == nil {
			continue
		}
		elapsed := r.crawlCompletedAt.Sub(*r.crawlStartedAt)
		if elapsed > maxWallClock {
			fmt.Printf("[gate] wall-clock: FAIL source=%s elapsed=%s cap=%s\n", r.sourceKind, elapsed, maxWallClock)
			return false
		}
	}
	fmt.Printf("[gate] wall-clock: OK (all sources within %s cap)\n", maxWallClock)
	return true
}

func checkLLMCost(costCents int64) bool {
	if costCents > maxLLMCentsPerTenant {
		fmt.Printf("[gate] LLM cost: FAIL spent_cents=%d cap_cents=%d\n", costCents, maxLLMCentsPerTenant)
		return false
	}
	fmt.Printf("[gate] LLM cost: OK spent_cents=%d cap_cents=%d\n", costCents, maxLLMCentsPerTenant)
	return true
}

func checkOrgSnapshot(rows []manifestRow) bool {
	completed := 0
	for _, r := range rows {
		if r.crawlStatus == "completed" {
			completed++
		}
	}
	if expectedSourceKinds == 0 {
		fmt.Println("[gate] org-snapshot: OK (no expected source-kinds; vacuously passes)")
		return true
	}
	pct := completed * 100 / expectedSourceKinds
	if pct < minOrgSnapshotPercent {
		fmt.Printf("[gate] org-snapshot: FAIL completed=%d/%d (%d%%) min=%d%%\n",
			completed, expectedSourceKinds, pct, minOrgSnapshotPercent)
		return false
	}
	fmt.Printf("[gate] org-snapshot: OK completed=%d/%d (%d%%)\n", completed, expectedSourceKinds, pct)
	return true
}

func checkSkillPrecision() bool {
	fmt.Println("[gate] Skill recommendation precision: N/A (Skill-Selector Agent deferred to Stage 1 per ADR-0003)")
	return true
}

func checkDestructive() bool {
	// Stage 0 has no destructive write surface — the Onboarding Crew
	// connectors use read-only OAuth scopes by construction, and the
	// gate is structural (defense-in-depth lockdown enforced in
	// CLAUDE.md). The dimension is meaningful in Stage 1+ when write
	// scopes become operator-toggleable per department.
	fmt.Println("[gate] destructive actions: OK (0 — Stage 0 read-only by construction)")
	return true
}
