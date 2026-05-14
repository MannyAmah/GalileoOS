// Package main is the Galileo API gateway entry point.
//
// Stage 0 / PR-B capability: JWT verification, Postgres-fresh tenant
// resolution, LiteLLM passthrough with metadata injection, budget cap
// enforcement, cost_events webhook ingestion, OpenTelemetry span
// emission. Configured entirely from environment variables.
//
// Environment variables (* = required):
//
//   GALILEO_GATEWAY_ADDR             listen address       default ":8080"
//   GALILEO_GATEWAY_PUBKEY           JWT public key path  default "kernel/auth/dev-keys/public.pem"
//   GALILEO_GATEWAY_DATABASE_URL *   Postgres DSN
//   GALILEO_GATEWAY_LITELLM_URL      LiteLLM base URL     default "http://localhost:4000"
//   GALILEO_GATEWAY_OTEL_ENDPOINT    OTel collector OTLP  default "localhost:4317"
//   GALILEO_COST_EVENTS_SECRET *     shared secret for LiteLLM webhook
//
// Boot sequence (each step exits the process on failure — fail-loud):
//
//   1. Open pgxpool
//   2. Run embedded migrations (kernel/cmd/gateway/migrations/)
//   3. Init OTel tracer provider
//   4. Wire Server with dependencies + start HTTP listen
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const serviceName = "galileo-gateway"

func main() {
	logger := log.New(os.Stderr, serviceName+" ", log.LstdFlags|log.LUTC)

	addr := envOr("GALILEO_GATEWAY_ADDR", ":8080")
	pubKeyPath := envOr("GALILEO_GATEWAY_PUBKEY", "kernel/auth/dev-keys/public.pem")
	dbURL := os.Getenv("GALILEO_GATEWAY_DATABASE_URL")
	if dbURL == "" {
		logger.Fatalf("GALILEO_GATEWAY_DATABASE_URL is required")
	}
	llmURL := envOr("GALILEO_GATEWAY_LITELLM_URL", "http://localhost:4000")
	otelEndpoint := envOr("GALILEO_GATEWAY_OTEL_ENDPOINT", "localhost:4317")
	costEventsSecret := os.Getenv("GALILEO_COST_EVENTS_SECRET")
	if costEventsSecret == "" {
		logger.Fatalf("GALILEO_COST_EVENTS_SECRET is required (shared with LiteLLM's GENERIC_LOGGER_HEADERS)")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		logger.Fatalf("postgres pool: %v", err)
	}
	defer pool.Close()

	migrationCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	if err := RunMigrations(migrationCtx, pool); err != nil {
		cancel()
		logger.Fatalf("migrations: %v", err)
	}
	cancel()

	otelShutdown, err := InitTracer(ctx, otelEndpoint)
	if err != nil {
		logger.Fatalf("otel tracer: %v", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = otelShutdown(shutdownCtx)
	}()

	tenants := NewTenantResolver(pool)
	llm, err := NewLiteLLMClient(llmURL)
	if err != nil {
		logger.Fatalf("litellm client: %v", err)
	}

	srv := NewServer(addr, pubKeyPath, costEventsSecret, tenants, llm, pool, logger)
	logger.Printf("listening on %s", addr)
	if err := srv.ListenAndServe(ctx); err != nil {
		logger.Fatalf("server: %v", err)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
