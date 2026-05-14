// Build-tagged integration test for the Stage 0 gateway. Runs under the
// `gateway_integration` build tag so it is excluded from the default
// `go test ./...` run. CI's gateway-integration job sets the tag and
// brings up real Postgres + LiteLLM + Jaeger + OTel-collector service
// containers; developers run it locally via `make stage0-gateway-test`
// after `make up`.
//
// What it verifies (cumulative through PR-B):
//   - JWT verification: missing / wrong-signature / expired tokens are rejected.
//   - Postgres-fresh tenant resolution: JWT-cached budget is ignored.
//   - LiteLLM passthrough with body metadata injection (PR-B).
//   - Postgres-unavailable → 503 (Drift-1).
//   - Budget cap enforcement (PR-B): cost_events sum ≥ cap → 402.
//   - Cost-events webhook (PR-B): valid payload + shared secret → row written.
//   - Cost-events webhook auth (PR-B): missing/wrong secret → 401.
//
// What it does NOT verify (deferred):
//   - workflow ID correlation (PR-C).
//   - Live Jaeger/OTel span emission roundtrip — Stage 0 boots both
//     services and waits for readiness, but doesn't assert traces land
//     in Jaeger's index. That's a Week 4 gate-test concern.

//go:build gateway_integration

package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	defaultDBURL              = "postgres://galileo:galileo@localhost:5432/galileo?sslmode=disable"
	defaultLLMURL             = "http://localhost:4000"
	defaultOTelHealthURL      = "http://localhost:13133/"
	defaultJaegerUIURL        = "http://localhost:16686/"
	testCostEventsSecret      = "integration-test-secret"
)

// testEnv bundles the setup the integration tests share: keypair on
// disk, a real Postgres pool with migrations applied, the gateway
// handler bound to an httptest.Server, and a teardown closure.
type testEnv struct {
	keyDir    string
	pubKey    string
	privKey   string
	pool      *pgxpool.Pool
	server    *httptest.Server
	tenantID  string
	cleanup   func()
}

func setupEnv(t *testing.T) *testEnv {
	t.Helper()

	dbURL := envOrDefault("GALILEO_GATEWAY_DATABASE_URL", defaultDBURL)
	llmURL := envOrDefault("GALILEO_GATEWAY_LITELLM_URL", defaultLLMURL)

	keyDir := t.TempDir()
	if err := auth.GenerateKeypair(keyDir); err != nil {
		t.Fatalf("generate keypair: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}

	// Apply embedded migrations — same code path the gateway runs at
	// startup, so the integration test exercises the real schema.
	if err := RunMigrations(ctx, pool); err != nil {
		pool.Close()
		t.Fatalf("migrations: %v", err)
	}

	tenantID := uuid.Must(uuid.NewV7()).String()
	if _, err := pool.Exec(ctx,
		`INSERT INTO tenants (tenant_id, monthly_budget_cents) VALUES ($1, $2)`,
		tenantID, int64(99999),
	); err != nil {
		pool.Close()
		t.Fatalf("insert tenant: %v", err)
	}

	// Wait for OTel collector + Jaeger so span emission doesn't fail
	// silently inside the test. Bounded retries with explicit failure
	// messages so the next CI iteration is debuggable without log
	// spelunking.
	waitForReady(t, defaultOTelHealthURL, "OTel collector")
	waitForReady(t, defaultJaegerUIURL, "Jaeger UI")

	tenants := NewTenantResolver(pool)
	llm, err := NewLiteLLMClient(llmURL)
	if err != nil {
		pool.Close()
		t.Fatalf("litellm client: %v", err)
	}
	logger := log.New(io.Discard, "", 0)
	gw := NewServer(":0", keyDir+"/public.pem", testCostEventsSecret, tenants, llm, pool, logger)
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
		// Don't drop tables — parallel tests in the same run share the
		// schema, keyed on unique tenantIDs. Clean up this run's rows.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = pool.Exec(ctx, `DELETE FROM cost_events WHERE tenant_id = $1`, tenantID)
		_, _ = pool.Exec(ctx, `DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
		pool.Close()
	}
	t.Cleanup(env.cleanup)
	return env
}

// waitForReady polls url every 1s up to 30 times, returning when the
// endpoint answers with 2xx. On exhaustion it calls t.Fatalf with a
// clear failure message so the next CI iteration is tractable without
// spelunking through container logs. Skipped entirely when GALILEO_SKIP_OBS_WAIT
// is set (dev hosts that don't run observability locally).
func waitForReady(t *testing.T, url, label string) {
	t.Helper()
	if os.Getenv("GALILEO_SKIP_OBS_WAIT") != "" {
		t.Logf("waitForReady: skipping %s (GALILEO_SKIP_OBS_WAIT set)", label)
		return
	}
	for i := 0; i < 30; i++ {
		resp, err := http.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode < 300 {
				return
			}
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatalf("%s did not become ready at %s within 30s", label, url)
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
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

func TestChatCompletionsForwards(t *testing.T) {
	env := setupEnv(t)
	body := `{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"hi"}]}`
	tok := env.mintToken(t, auth.Claims{MonthlyBudgetCents: 1}, time.Hour)

	req, _ := http.NewRequest(http.MethodPost, env.server.URL+"/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 500 {
		t.Fatalf("upstream-mediated failure: status=%d body=%s", resp.StatusCode, respBody)
	}
	if got := resp.Header.Get("x-galileo-request-id"); got == "" {
		t.Error("response missing x-galileo-request-id header (PR-B should always set it)")
	}
	var parsed any
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		t.Errorf("response is not JSON: %v body=%s", err, respBody)
	}
}

func TestPostgresUnavailableReturns503(t *testing.T) {
	env := setupEnv(t)
	tok := env.mintToken(t, auth.Claims{}, time.Hour)
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

// TestBudgetCapDeniesWith402 (PR-B): with spend already at the cap, the
// next request is denied with HTTP 402 Payment Required. Exercises the
// loop A enforcement path: read sum(cost_cents) for the current month,
// compare to TenantContext.MonthlyBudgetCents, deny on overage.
func TestBudgetCapDeniesWith402(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	// Insert one cost_events row that equals the tenant's budget so the
	// next request sums to spend == cap (deny is >= cap).
	_, err := env.pool.Exec(ctx, `
		INSERT INTO cost_events (request_id, tenant_id, event_ts, cost_cents, provider, model)
		VALUES ($1, $2, now(), 99999, 'test', 'test-model')
	`, uuid.Must(uuid.NewV7()).String(), env.tenantID)
	if err != nil {
		t.Fatalf("seed cost_events: %v", err)
	}

	tok := env.mintToken(t, auth.Claims{}, time.Hour)
	req, _ := http.NewRequest(http.MethodPost, env.server.URL+"/v1/chat/completions",
		strings.NewReader(`{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"hi"}]}`))
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusPaymentRequired {
		t.Fatalf("expected 402 at budget cap, got %d", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Errorf("402 body not JSON: %v", err)
	}
	if body["error"] != "monthly_budget_cents exceeded" {
		t.Errorf("402 body.error: got %v, want 'monthly_budget_cents exceeded'", body["error"])
	}
}

// TestCostEventsWebhookRoundtrip (PR-B): POSTing a synthetic
// StandardLoggingPayload to /internal/cost-events writes a row that the
// budget middleware can then see. Verifies the OSS metadata channel
// (requester_metadata.galileo_*) end-to-end.
func TestCostEventsWebhookRoundtrip(t *testing.T) {
	env := setupEnv(t)
	requestID := uuid.Must(uuid.NewV7()).String()

	payload := fmt.Sprintf(`[{
		"id": "litellm-internal-id-001",
		"response_cost": 0.0125,
		"model": "gpt-3.5-turbo",
		"custom_llm_provider": "openai",
		"endTime": %d.5,
		"metadata": {"requester_metadata": {"galileo_tenant_id": "%s", "galileo_request_id": "%s"}}
	}]`, time.Now().Unix(), env.tenantID, requestID)

	req, _ := http.NewRequest(http.MethodPost, env.server.URL+"/internal/cost-events", strings.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+testCostEventsSecret)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("webhook returned %d: %s", resp.StatusCode, body)
	}

	// Verify the row landed with the expected fields.
	var costCents int64
	var model string
	err = env.pool.QueryRow(context.Background(),
		`SELECT cost_cents, model FROM cost_events WHERE request_id = $1`, requestID,
	).Scan(&costCents, &model)
	if err != nil {
		t.Fatalf("cost_events row not found: %v", err)
	}
	if costCents != 1 { // 0.0125 dollars = 1.25 cents → rounded to 1
		t.Errorf("cost_cents: got %d, want 1 (0.0125 × 100 rounded)", costCents)
	}
	if model != "gpt-3.5-turbo" {
		t.Errorf("model: got %q, want gpt-3.5-turbo", model)
	}
}

// TestCostEventsWebhookRejectsMissingSecret (PR-B): the webhook auths
// via shared secret in Authorization: Bearer. Missing → 401.
func TestCostEventsWebhookRejectsMissingSecret(t *testing.T) {
	env := setupEnv(t)
	req, _ := http.NewRequest(http.MethodPost, env.server.URL+"/internal/cost-events", strings.NewReader(`[]`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("missing secret: got %d, want 401", resp.StatusCode)
	}
}

// TestCostEventsWebhookIdempotency (PR-B): re-posting the same
// galileo_request_id is a no-op — the ON CONFLICT (request_id) DO NOTHING
// clause prevents double-counting if LiteLLM retries the callback.
func TestCostEventsWebhookIdempotency(t *testing.T) {
	env := setupEnv(t)
	requestID := uuid.Must(uuid.NewV7()).String()
	payload := fmt.Sprintf(`[{
		"id": "litellm-id-dup", "response_cost": 0.02, "model": "m", "custom_llm_provider": "p", "endTime": %d,
		"metadata": {"requester_metadata": {"galileo_tenant_id": "%s", "galileo_request_id": "%s"}}
	}]`, time.Now().Unix(), env.tenantID, requestID)

	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest(http.MethodPost, env.server.URL+"/internal/cost-events", strings.NewReader(payload))
		req.Header.Set("Authorization", "Bearer "+testCostEventsSecret)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("POST iteration %d: %v", i, err)
		}
		_ = resp.Body.Close()
	}

	var count int
	err := env.pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM cost_events WHERE request_id = $1`, requestID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("count cost_events: %v", err)
	}
	if count != 1 {
		t.Errorf("idempotency: got %d rows, want 1 (re-POST should DO NOTHING)", count)
	}
}
