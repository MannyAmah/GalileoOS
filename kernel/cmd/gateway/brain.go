// Brain extension verification — boot-time fail-loud check that the
// Brain substrate's two required extensions (pgvector + AGE) are
// present in the live database. Belt-and-suspenders against the
// failure mode where a migration thinks it ran but the extension is
// somehow missing (e.g., an operator manually dropped it). Same
// fail-loud posture as PR-A's Postgres-unavailable check.

package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// verifyExtensions runs after RunMigrations during gateway boot. The
// migration in 0005_brain.sql CREATE EXTENSIONs both; this function
// reads pg_extension to confirm. If either is missing, gateway exits
// loudly with a named-likely-cause error message — same pattern as
// PR-A's connection-error surface.
func verifyExtensions(ctx context.Context, pool *pgxpool.Pool) error {
	var count int
	err := pool.QueryRow(ctx,
		"SELECT count(*) FROM pg_extension WHERE extname IN ('vector', 'age')",
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("checking brain extensions in pg_extension: %w", err)
	}
	if count != 2 {
		return fmt.Errorf(
			"brain extensions missing (expected 2 of vector + age, got %d) — "+
				"likely cause: extension was manually dropped, or the postgres-brain "+
				"image isn't in use; re-run migrations or CREATE EXTENSION manually, "+
				"see deploy/compose/postgres-brain/README.md",
			count,
		)
	}
	return nil
}
