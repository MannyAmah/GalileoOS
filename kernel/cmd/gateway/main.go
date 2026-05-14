// Package main is the Galileo API gateway entry point.
//
// Stage 0 / PR-A: HTTP server with JWT verification (Ed25519 dev keypair),
// Postgres-fresh tenant resolution, and a LiteLLM passthrough. Configured
// entirely from environment variables so the same binary can run from
// the developer's host (against `make up` compose services) or from CI's
// service-container job.
//
// Environment variables (all required unless noted):
//
//	GALILEO_GATEWAY_ADDR        Listen address          default ":8080"
//	GALILEO_GATEWAY_PUBKEY      Path to JWT public key  default "kernel/auth/dev-keys/public.pem"
//	GALILEO_GATEWAY_DATABASE_URL  Postgres DSN          (no default)
//	GALILEO_GATEWAY_LITELLM_URL   LiteLLM base URL      default "http://localhost:4000"
//
// Wiring scope deferred to later PRs is enumerated in server.go.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		logger.Fatalf("postgres pool: %v", err)
	}
	defer pool.Close()

	tenants := NewTenantResolver(pool)
	llm, err := NewLiteLLMClient(llmURL)
	if err != nil {
		logger.Fatalf("litellm client: %v", err)
	}

	srv := NewServer(addr, pubKeyPath, tenants, llm, logger)
	fmt.Fprintf(os.Stdout, "%s listening on %s\n", serviceName, addr)
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
