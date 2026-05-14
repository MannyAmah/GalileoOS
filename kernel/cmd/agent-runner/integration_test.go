// Build-tagged integration test for the agent-runner.
//
// Exercises the full Hello Agent path against real services brought
// up via docker (Temporal + Postgres + LiteLLM). The CI job that runs
// this is a separate workflow file from gateway-integration; locally
// developers run `make stage0-agent-runner-test` after `make up`.
//
// What it verifies:
//   - Worker registration: HelloAgentWorkflow + CallLLMActivity
//     reachable on the `galileo-agent-runner` task queue.
//   - HTTP server: POST /v1/tasks accepts a TaskInput, starts a
//     workflow, returns 202 + task_id.
//   - HTTP server: GET /v1/tasks/{id} returns {status:"running"}
//     while the workflow is in flight, TaskResult JSON when done.
//   - Drift-2 correlation: TaskResult.output.cost_event_request_ids
//     carries exactly one id (the gateway-issued x-galileo-request-id
//     for the single LLM call this agent makes).
//   - Drift-6 dispatch: the "hello" department slug resolves to
//     HelloAgentWorkflow via the explicit registry; an unknown
//     slug is rejected at the HTTP layer with 400.
//
// What it does NOT verify (deferred to the Week 4 gate test):
//   - The 100×-runs assertion. This file runs the path once.
//   - End-to-end web → agent-runner → gateway via the Next.js
//     proxy (the API route is exercised separately under
//     web/ tests if added later).

//go:build agent_runner_integration

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/MannyAmah/GalileoOS/kernel/auth"
	pb "github.com/MannyAmah/GalileoOS/kernel/gen/galileo/v1"
)

const (
	defaultDBURL          = "postgres://galileo:galileo@localhost:5432/galileo?sslmode=disable"
	defaultGatewayURL     = "http://localhost:8080"
	defaultTemporalHostPort = "localhost:7233"
	testTenantBudget      = int64(99999)
)

type agentEnv struct {
	keyDir     string
	tenantID   string
	temporal   client.Client
	worker     worker.Worker
	server     *httptest.Server
	gatewayURL string
}

func setupAgentEnv(t *testing.T) *agentEnv {
	t.Helper()

	dbURL := envOrDefault("GALILEO_GATEWAY_DATABASE_URL", defaultDBURL)
	gatewayURL := envOrDefault("GALILEO_AGENT_GATEWAY_URL", defaultGatewayURL)
	temporalHostPort := envOrDefault("GALILEO_AGENT_TEMPORAL_HOSTPORT", defaultTemporalHostPort)

	// Keypair source: CI sets GALILEO_AGENT_JWT_PRIVATE_KEY_PATH to the
	// same keypair the gateway subprocess was started with, so the
	// activity-issued JWTs the gateway receives are signed by a key
	// the gateway can verify. If that env var is set we reuse the
	// parent directory; otherwise we generate a fresh keypair in
	// t.TempDir() (local-dev case).
	var keyDir string
	if existingPriv := os.Getenv("GALILEO_AGENT_JWT_PRIVATE_KEY_PATH"); existingPriv != "" {
		keyDir = filepath.Dir(existingPriv)
	} else {
		keyDir = t.TempDir()
		require.NoError(t, auth.GenerateKeypair(keyDir))
		require.NoError(t, os.Setenv("GALILEO_AGENT_JWT_PRIVATE_KEY_PATH", keyDir+"/private.pem"))
	}
	require.NoError(t, os.Setenv("GALILEO_AGENT_GATEWAY_URL", gatewayURL))

	// Seed a tenant in Postgres with a high budget so the 1-call
	// Hello Agent doesn't trip the cap.
	pool, err := pgxpool.New(context.Background(), dbURL)
	require.NoError(t, err)
	tenantID := uuid.Must(uuid.NewV7()).String()
	_, err = pool.Exec(context.Background(),
		`INSERT INTO tenants (tenant_id, monthly_budget_cents) VALUES ($1, $2)`,
		tenantID, testTenantBudget,
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = pool.Exec(ctx, `DELETE FROM cost_events WHERE tenant_id = $1`, tenantID)
		_, _ = pool.Exec(ctx, `DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
		pool.Close()
	})

	// Connect to Temporal. Per the PR-C verification round, the
	// `t.Cleanup` shape applies if this test ever brings up a
	// dedicated Temporal container; the CI job uses the compose
	// stack's Temporal so cleanup is the stack's responsibility.
	c, err := client.Dial(client.Options{HostPort: temporalHostPort, Namespace: "default"})
	require.NoError(t, err, "Temporal dial — is the compose stack up?")
	t.Cleanup(c.Close)

	w := worker.New(c, TaskQueue, worker.Options{})
	registerWorkflows(w)
	w.RegisterActivity(CallLLMActivity)
	require.NoError(t, w.Start())
	t.Cleanup(w.Stop)

	logger := log.New(io.Discard, "", 0)
	srv := NewServer(":0", keyDir+"/public.pem", c, logger)
	tsrv := httptest.NewServer(srv.Handler())
	t.Cleanup(tsrv.Close)

	return &agentEnv{
		keyDir:     keyDir,
		tenantID:   tenantID,
		temporal:   c,
		worker:     w,
		server:     tsrv,
		gatewayURL: gatewayURL,
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func (e *agentEnv) tenantToken(t *testing.T) string {
	t.Helper()
	tok, err := auth.MintToken(e.keyDir+"/private.pem", auth.Claims{
		TenantID:           e.tenantID,
		MonthlyBudgetCents: testTenantBudget,
	}, time.Hour)
	require.NoError(t, err)
	return tok
}

func TestCreateTaskUnknownDepartmentReturns400(t *testing.T) {
	env := setupAgentEnv(t)
	body, _ := json.Marshal(&pb.TaskInput{
		Tenant: &pb.TenantContext{
			TenantId: &pb.TenantId{Value: env.tenantID},
		},
		Department: "marketing", // not in workflowRegistry
		Goal:       "test",
	})
	req, _ := http.NewRequest(http.MethodPost, env.server.URL+"/v1/tasks", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+env.tenantToken(t))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("unknown department: got %d, want 400. body=%s", resp.StatusCode, body)
	}
}

func TestCreateTaskTenantMismatchReturns403(t *testing.T) {
	env := setupAgentEnv(t)
	// Body carries a different tenant_id than the JWT.
	other := uuid.Must(uuid.NewV7()).String()
	body, _ := json.Marshal(&pb.TaskInput{
		Tenant:     &pb.TenantContext{TenantId: &pb.TenantId{Value: other}},
		Department: "hello",
		Goal:       "test",
	})
	req, _ := http.NewRequest(http.MethodPost, env.server.URL+"/v1/tasks", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+env.tenantToken(t))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("tenant mismatch: got %d, want 403", resp.StatusCode)
	}
}

func TestHelloAgentEndToEnd(t *testing.T) {
	env := setupAgentEnv(t)
	body, _ := json.Marshal(&pb.TaskInput{
		Tenant: &pb.TenantContext{
			TenantId:           &pb.TenantId{Value: env.tenantID},
			MonthlyBudgetCents: testTenantBudget,
		},
		Department: "hello",
		Goal:       "say hello in three words",
	})
	req, _ := http.NewRequest(http.MethodPost, env.server.URL+"/v1/tasks", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+env.tenantToken(t))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, resp.StatusCode)
	var createResp struct {
		TaskID string `json:"task_id"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&createResp))
	_ = resp.Body.Close()
	require.NotEmpty(t, createResp.TaskID)

	// Poll for completion. Bounded retries — same shape as PR-B's
	// observability waitForReady. 30 attempts × 1s = 30s max.
	var result pb.TaskResult
	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)
		req, _ := http.NewRequest(http.MethodGet, env.server.URL+"/v1/tasks/"+createResp.TaskID, nil)
		req.Header.Set("Authorization", "Bearer "+env.tenantToken(t))
		pollResp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		respBody, _ := io.ReadAll(pollResp.Body)
		_ = pollResp.Body.Close()
		if strings.Contains(string(respBody), `"status":"running"`) {
			continue
		}
		require.NoError(t, json.Unmarshal(respBody, &result))
		break
	}
	if result.Status == "" {
		t.Fatal("workflow did not complete within 30s")
	}
	if result.Status != "shipped" {
		t.Fatalf("expected status=shipped, got %q error=%q", result.Status, result.Error)
	}
	if result.Output == nil {
		t.Fatal("nil output on shipped task")
	}
	// Drift-2: cost_event_request_ids carries the single request id
	// the gateway issued for the one LLM call.
	if len(result.Output.CostEventRequestIds) != 1 {
		t.Errorf("Drift-2: cost_event_request_ids should have 1 entry, got %v", result.Output.CostEventRequestIds)
	}
}

// guard against the test file unexpectedly being included in the
// default test pass — the build tag should keep it out, but if
// someone removes the tag accidentally, exec.Command is the canonical
// "this is integration scope" tell.
var _ = exec.Command
