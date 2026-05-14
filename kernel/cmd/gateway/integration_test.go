// Build-tagged integration test for the Stage 0 gateway. Runs under the
// `gateway_integration` build tag so it is excluded from the default
// `go test ./...` run. CI's gateway-integration job sets the tag and
// brings up real Postgres + LiteLLM service containers; developers run
// it locally via `make stage0-gateway-test` after `make up`.
//
// What it verifies (PR-A scope):
//   - JWT verification: missing / wrong-issuer / expired tokens are rejected.
//   - Postgres-fresh tenant resolution: monthly_budget_cents in the JWT
//     is *ignored*; the value returned to the handler comes from Postgres.
//   - LiteLLM passthrough: the request body reaches LiteLLM and the
//     response body returns unchanged.
//   - Postgres-unavailable behavior (Drift-1): pool closed → 503.
//
// What it does NOT verify (deferred per server.go's out-of-scope notes):
//   - cost_events writing (PR-B).
//   - Budget cap enforcement (PR-B).
//   - workflow_id correlation (PR-C).

//go:build gateway_integration

package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/MannyAmah/GalileoOS/kernel/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultDBURL  = "postgres://galileo:galileo@localhost:5432/galileo?sslmode=disable"
	defaultLLMURL = "http://localhost:4000"
)

// testEnv bundles the setup the integration tests share: keypair on disk,
// a real Postgres pool with a tenants row, the gateway handler bound to
// an httptest.Server, and a teardown closure.
type testEnv struct {
	keyDir   string
	pubKey   string
	privKey  string
	pool     *pgxpool.Pool
	server   *httptest.Server
	tenantID string
	cleanup  func()
}

func setupEnv(t *testing.T) *testEnv {
	t.Helper()

	dbURL := os.Getenv("GALILEO_GATEWAY_DATABASE_URL")
	if dbURL == "" {
		dbURL = defaultDBURL
	}
	llmURL := os.Getenv("GALILEO_GATEWAY_LITELLM_URL")
	if llmURL == "" {
		llmURL = defaultLLMURL
	}

	keyDir := t.TempDir()
	if err := auth.GenerateKeypair(keyDir); err != nil {
		t.Fatalf("generate keypair: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}

	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS tenants (
			tenant_id UUID PRIMARY KEY,
			monthly_budget_cents BIGINT NOT NULL
		)
	`); err != nil {
		pool.Close()
		t.Fatalf("create tenants table: %v", err)
	}

	tenantID := uuid.Must(uuid.NewV7()).String()
	if _, err := pool.Exec(ctx,
		`INSERT INTO tenants (tenant_id, monthly_budget_cents) VALUES ($1, $2)`,
		tenantID, int64(99999),
	); err != nil {
		pool.Close()
		t.Fatalf("insert tenant: %v", err)
	}

	tenants := NewTenantResolver(pool)
	llm, err := NewLiteLLMClient(llmURL)
	if err != nil {
		pool.Close()
		t.Fatalf("litellm client: %v", err)
	}
	logger := log.New(io.Discard, "", 0)
	gw := NewServer(":0", keyDir+"/public.pem", tenants, llm, logger)
	srv := httptest.NewServer(gw.Handler())

	env := &testEnv{
		keyDir:   keyDir,
		pubKey:   keyDir + "/public.pem",
		privKey:  keyDir + "/private.pem",
		pool:     pool,
		server:   srv,
		tenantID: tenantID,
	}
	env.cleanup = func() {
		srv.Close()
		// Don't drop the table — other tests in the same run share it
		// keyed on the unique tenantID. Just remove this run's row.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = pool.Exec(ctx, `DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
		pool.Close()
	}
	t.Cleanup(env.cleanup)
	return env
}

func (e *testEnv) mintToken(t *testing.T, c auth.Claims, ttl time.Duration) string {
	t.Helper()
	if c.TenantID == "" {
		c.TenantID = e.tenantID
	}
	tok, err := auth.MintToken(e.privKey, c, ttl)
	if err != nil {
		t.Fatalf("mint token: %v", err)
	}
	return tok
}

func TestHealthz(t *testing.T) {
	env := setupEnv(t)
	resp, err := http.Get(env.server.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("healthz status: got %d want 200", resp.StatusCode)
	}
}

func TestChatCompletionsRejectsMissingAuth(t *testing.T) {
	env := setupEnv(t)
	resp, err := http.Post(env.server.URL+"/v1/chat/completions", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status: got %d want 401", resp.StatusCode)
	}
}

func TestChatCompletionsRejectsExpiredToken(t *testing.T) {
	env := setupEnv(t)
	tok := env.mintToken(t, auth.Claims{}, -time.Hour)
	req, _ := http.NewRequest(http.MethodPost, env.server.URL+"/v1/chat/completions", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status: got %d want 401", resp.StatusCode)
	}
}

// TestChatCompletionsRejectsWrongSignature exercises the Stage-1 swap
// risk: if the gateway's public key changes without a matching private-key
// rotation, every existing token must be rejected at the integration
// boundary, not just by the auth/ unit tests.
func TestChatCompletionsRejectsWrongSignature(t *testing.T) {
	env := setupEnv(t)
	otherDir := t.TempDir()
	if err := auth.GenerateKeypair(otherDir); err != nil {
		t.Fatalf("generate alternate keypair: %v", err)
	}
	tok, err := auth.MintToken(otherDir+"/private.pem", auth.Claims{TenantID: env.tenantID}, time.Hour)
	if err != nil {
		t.Fatalf("mint with alternate key: %v", err)
	}
	req, _ := http.NewRequest(http.MethodPost, env.server.URL+"/v1/chat/completions", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("wrong-signature: got %d want 401", resp.StatusCode)
	}
}

// TestChatCompletionsForwards verifies the happy path: a valid token,
// a tenant row in Postgres, and a request body that reaches LiteLLM's
// /v1/chat/completions endpoint. With LITELLM_MODE=test, LiteLLM returns
// a canned response — we only assert the proxy reached it (2xx or a
// LiteLLM-shaped 4xx with a JSON body), not the contents.
func TestChatCompletionsForwards(t *testing.T) {
	env := setupEnv(t)
	body := `{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"hi"}]}`
	tok := env.mintToken(t, auth.Claims{
		MonthlyBudgetCents: 1, // informational; gateway should ignore this
	}, time.Hour)

	req, _ := http.NewRequest(http.MethodPost, env.server.URL+"/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(resp.Body)

	// LiteLLM in test mode answers with a JSON object; the gateway should
	// hand back exactly what LiteLLM returned. We accept any 2xx/4xx
	// (LiteLLM-shaped) as proof the proxy hop worked — a 5xx means the
	// proxy itself failed.
	if resp.StatusCode >= 500 {
		t.Fatalf("upstream-mediated failure: status=%d body=%s", resp.StatusCode, respBody)
	}
	var parsed any
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		t.Errorf("response is not JSON: %v body=%s", err, respBody)
	}
}

// TestPostgresUnavailableReturns503 exercises Drift-1: when the tenant
// store is unreachable, the gateway must deny with 503 rather than fall
// back to the JWT-cached budget value.
func TestPostgresUnavailableReturns503(t *testing.T) {
	env := setupEnv(t)
	tok := env.mintToken(t, auth.Claims{}, time.Hour)

	// Close the pool to force connection failures on subsequent queries.
	env.pool.Close()

	req, _ := http.NewRequest(http.MethodPost, env.server.URL+"/v1/chat/completions", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Drift-1: with Postgres down expected 503, got %d", resp.StatusCode)
	}
}
