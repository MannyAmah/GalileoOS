// CallLLMActivity — the only activity in the Hello Agent workflow.
//
// What it does:
//   1. Mints a fresh JWT for the tenant carried in TaskInput, using
//      the Ed25519 dev keypair at GALILEO_AGENT_JWT_PRIVATE_KEY_PATH.
//      (Stage 1 swaps this for a Supabase service-account token at the
//      same env-var surface.)
//   2. POSTs an OpenAI-format chat-completion request to the gateway,
//      with the TaskInput.goal as the user message.
//   3. Reads the gateway's x-galileo-request-id response header — this
//      is the cost_events.request_id (Drift-2 correlation).
//   4. Parses the LiteLLM response for the assistant message text and
//      returns an AgentOutput with body, cost_cents (computed from
//      response usage; Stage 0 echoes the LiteLLM response_cost back),
//      and cost_event_request_ids carrying the gateway's id (the
//      Drift-2 field's first consumer).
//
// Budget-cap handling: the gateway returns 402 on overage. The activity
// surfaces that as a non-retryable error so the workflow returns
// "rejected" cleanly to the web — retries on a budget-denied tenant
// would just deny again.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"time"

	"go.temporal.io/sdk/temporal"

	"github.com/MannyAmah/GalileoOS/kernel/auth"
	pb "github.com/MannyAmah/GalileoOS/kernel/gen/galileo/v1"
)

// activityHTTP is the http.Client the activity uses. Package-level so
// integration tests can swap it for a recording client; production
// gets the default 60s timeout matching the gateway's own LiteLLM
// timeout (one-hop budget).
var activityHTTP = &http.Client{Timeout: 60 * time.Second}

// CallLLMActivity is the Temporal activity that talks to the gateway.
// Plain context.Context (not workflow.Context — activities run on the
// worker, not the workflow scheduler).
func CallLLMActivity(ctx context.Context, input *pb.TaskInput) (*pb.AgentOutput, error) {
	gwURL := envOr("GALILEO_AGENT_GATEWAY_URL", "http://localhost:8080")
	privKeyPath := envOr("GALILEO_AGENT_JWT_PRIVATE_KEY_PATH", "kernel/auth/dev-keys/private.pem")

	tenantID := tenantIDFromInput(input)
	if tenantID == "" {
		return nil, temporal.NewNonRetryableApplicationError("missing tenant_id on TaskInput", "BadRequest", nil)
	}

	token, err := mintTenantToken(privKeyPath, tenantID, input.GetTenant().GetMonthlyBudgetCents())
	if err != nil {
		// JWT minting failures are environmental (missing key, bad
		// permissions) — retryable so Temporal redrives once the
		// operator fixes the deployment.
		return nil, fmt.Errorf("mint tenant jwt: %w", err)
	}

	body := chatCompletionsBody(input.GetGoal())
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal chat body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, gwURL+"/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("build gateway request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := activityHTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gateway request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read gateway response: %w", err)
	}

	// 402 = budget cap exceeded. Non-retryable — retrying would deny
	// again. Workflow translates this to TaskResult.status="rejected".
	if resp.StatusCode == http.StatusPaymentRequired {
		return nil, temporal.NewNonRetryableApplicationError(
			"monthly_budget_cents exceeded for tenant",
			"BudgetExceeded",
			errors.New(string(respBody)),
		)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("gateway returned %d: %s", resp.StatusCode, respBody)
	}

	requestID := resp.Header.Get("x-galileo-request-id")
	if requestID == "" {
		return nil, fmt.Errorf("gateway did not set x-galileo-request-id (PR-B regression?)")
	}

	body0, err := parseChatCompletionResponse(respBody)
	if err != nil {
		return nil, fmt.Errorf("parse chat response: %w", err)
	}

	return &pb.AgentOutput{
		Body:                body0.message,
		CostCents:           dollarsToCents(body0.responseCost),
		CostEventRequestIds: []string{requestID},
	}, nil
}

// tenantIDFromInput extracts the tenant UUID string from the
// TaskInput's TenantContext.tenant_id wrapper. Returns "" if missing.
func tenantIDFromInput(input *pb.TaskInput) string {
	if input == nil || input.GetTenant() == nil || input.GetTenant().GetTenantId() == nil {
		return ""
	}
	return input.GetTenant().GetTenantId().GetValue()
}

// mintTenantToken signs a short-lived JWT for tenantID using the
// Ed25519 private key at privKeyPath. The token's budget claim is
// informational (gateway re-reads from Postgres per Drift-1).
func mintTenantToken(privKeyPath, tenantID string, budgetCents int64) (string, error) {
	tok, err := auth.MintToken(privKeyPath, auth.Claims{
		TenantID:           tenantID,
		MonthlyBudgetCents: budgetCents,
	}, 10*time.Minute)
	if err != nil {
		// Strip the key path from the error so it doesn't end up in
		// activity logs — Drift-3 keypair is gitignored but the path
		// itself is unnecessary log surface.
		return "", errors.New("mint failed (check GALILEO_AGENT_JWT_PRIVATE_KEY_PATH)")
	}
	return tok, nil
}

// chatCompletionsBody builds an OpenAI-format chat completion request
// for the goal. Stage 0 demo uses gpt-3.5-turbo with one user message;
// Stage 1 grows to per-department system prompts and tool choice.
func chatCompletionsBody(goal string) map[string]any {
	return map[string]any{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "user", "content": goal},
		},
	}
}

// chatCompletionParsed pulls the assistant message text + the
// LiteLLM-reported cost (in dollars) from a chat-completion response
// body. response_cost is a LiteLLM extension to the OpenAI schema; if
// absent (operator-direct calls to an upstream OpenAI without
// LiteLLM), cost_cents will be 0 and the recon path is unaffected
// because cost_events is written by LiteLLM's callback to the gateway.
type chatCompletionParsed struct {
	message      string
	responseCost float64
}

func parseChatCompletionResponse(body []byte) (chatCompletionParsed, error) {
	var raw struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		ResponseCost float64 `json:"response_cost"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return chatCompletionParsed{}, fmt.Errorf("json: %w", err)
	}
	if len(raw.Choices) == 0 {
		return chatCompletionParsed{}, errors.New("no choices in response")
	}
	return chatCompletionParsed{
		message:      raw.Choices[0].Message.Content,
		responseCost: raw.ResponseCost,
	}, nil
}

func dollarsToCents(d float64) int64 {
	if d < 0 {
		return 0
	}
	return int64(math.Round(d * 100))
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
