// Stage 0 inline migration runner.
//
// Four structural pieces, each is load-bearing:
//
//  1. schema_migrations table creation is the runner's own first step.
//     A fresh database has no idea what "applied" means; we bootstrap
//     that before reading anything else. Idempotent — CREATE IF NOT
//     EXISTS — so existing databases skip it.
//
//  2. Migration filenames use strict naming: NNNN_description.sql with a
//     zero-padded sequence number. The runner sorts lexicographically and
//     applies in order. Any filename that doesn't match halts the runner
//     with a parse error — silent skipping is the failure mode we are
//     specifically refusing.
//
//  3. Each migration runs inside a transaction. The SQL body executes
//     first; the INSERT INTO schema_migrations(version) row commits in
//     the same transaction. A crash partway through can never leave a
//     half-applied migration marked as complete — either both the DDL
//     and the bookkeeping row commit, or neither does.
//
//  4. The runner is invoked at gateway startup, before HTTP listen. If
//     migrations fail, the process exits with the offending filename and
//     the database error. Same fail-loud posture as Postgres-unavailable
//     in PR-A — we never serve traffic against a half-migrated database.
//
// When the inline runner stops being sufficient (see CLAUDE.md "Migration
// tooling"), evaluate golang-migrate/migrate. The triggers are
// structural, not migration-count.

package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed all:migrations
var embeddedMigrations embed.FS

var migrationFilenameRE = regexp.MustCompile(`^(\d{4})_[a-z0-9_]+\.sql$`)

// RunMigrations applies every embedded migration not already in
// schema_migrations, in lexicographic filename order, each inside its
// own transaction. Safe to call repeatedly; only un-applied migrations
// execute.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    INT         PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	applied, err := loadAppliedVersions(ctx, pool)
	if err != nil {
		return err
	}

	files, err := listMigrationFiles()
	if err != nil {
		return err
	}

	for _, f := range files {
		version, err := parseMigrationVersion(f)
		if err != nil {
			return err
		}
		if _, ok := applied[version]; ok {
			continue
		}
		body, err := embeddedMigrations.ReadFile("migrations/" + f)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", f, err)
		}
		if err := applyMigration(ctx, pool, version, f, string(body)); err != nil {
			return fmt.Errorf("apply migration %s: %w", f, err)
		}
	}
	return nil
}

func loadAppliedVersions(ctx context.Context, pool *pgxpool.Pool) (map[int]struct{}, error) {
	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("read schema_migrations: %w", err)
	}
	defer rows.Close()
	applied := map[int]struct{}{}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("scan schema_migrations row: %w", err)
		}
		applied[v] = struct{}{}
	}
	return applied, rows.Err()
}

func listMigrationFiles() ([]string, error) {
	entries, err := fs.ReadDir(embeddedMigrations, "migrations")
	if err != nil {
		return nil, fmt.Errorf("read embedded migrations: %w", err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !migrationFilenameRE.MatchString(name) {
			return nil, fmt.Errorf("migrations: unexpected filename %q (expected NNNN_description.sql)", name)
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

func parseMigrationVersion(filename string) (int, error) {
	m := migrationFilenameRE.FindStringSubmatch(filename)
	if m == nil {
		return 0, fmt.Errorf("migrations: filename %q does not match NNNN_description.sql", filename)
	}
	n, err := strconv.Atoi(strings.TrimLeft(m[1], "0"))
	if err != nil {
		// Empty after trim means version 0 — disallow; sequence starts at 1.
		if m[1] == "0000" {
			return 0, fmt.Errorf("migrations: filename %q has version 0000; versions start at 0001", filename)
		}
		return 0, fmt.Errorf("migrations: parse version of %q: %w", filename, err)
	}
	return n, nil
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, version int, filename, body string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, body); err != nil {
		return fmt.Errorf("execute %s: %w", filename, err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
		return fmt.Errorf("record %s: %w", filename, err)
	}
	return tx.Commit(ctx)
}
