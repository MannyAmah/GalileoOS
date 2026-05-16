// Cost-events webhook receiver (cost-meter loop B).
//
// LiteLLM's OSS generic_api callback POSTs a List[StandardLoggingPayload]
// to GENERIC_LOGGER_ENDPOINT on every successful LLM call. We point that
// endpoint at /internal/cost-events on this gateway and authenticate
// with a shared secret in GENERIC_LOGGER_HEADERS — not a JWT, because
// the caller is the LiteLLM container, not a tenant.
//
// The handler reads the gateway-injected metadata
// (requester_metadata.galileo_tenant_id, .galileo_request_id) off each
// payload entry, converts response_cost (float dollars) to cost_cents
// BIGINT via ×100 with rounding, and INSERTs into cost_events with
// ON CONFLICT (request_id) DO NOTHING for idempotency.
//
// Payload entries that lack our metadata fields are skipped with a log
// line — they correspond to calls that bypassed the gateway entirely
// (e.g., direct LiteLLM /chat/completions from an unrelated client).
// Stage 0 deployment doesn't expose LiteLLM externally so this is
// belt-and-suspenders.

package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

// liteLLMPayload mirrors the fields we read out of
// StandardLoggingPayload. Other fields (response, messages, hidden_params,
// etc.) are accepted as JSON but ignored — we only need cost + metadata.
type liteLLMPayload struct {
	ID                  string                 `json:"id"`            // LiteLLM call id (informational)
	ResponseCost        float64                `json:"response_cost"` // dollars
	Model               string                 `json:"model"`
	CustomLLMProvider   string                 `json:"custom_llm_provider"`
	EndTime             float64                `json:"endTime"` // Unix seconds (float)
	Metadata            liteLLMPayloadMetadata `json:"metadata"`
}

type liteLLMPayloadMetadata struct {
	RequesterMetadata map[string]string `json:"requester_metadata"`
}

// costEventsHandler ingests one or more StandardLoggingPayloads from
// LiteLLM and writes them to cost_events idempotently.
func (s *Server) costEventsHandler(w http.ResponseWriter, r *http.Request) {
	if !s.checkCostEventsAuth(r) {
		writeErr(w, http.StatusUnauthorized, "missing or invalid shared secret")
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "read body: "+err.Error())
		return
	}
	var payloads []liteLLMPayload
	if err := json.Unmarshal(body, &payloads); err != nil {
		// LiteLLM may send a single object or a list depending on batch
		// settings — accept both shapes.
		var single liteLLMPayload
		if err2 := json.Unmarshal(body, &single); err2 != nil {
			writeErr(w, http.StatusBadRequest, "parse payload: "+err.Error())
			return
		}
		payloads = []liteLLMPayload{single}
	}

	written := 0
	for _, p := range payloads {
		tenantID := p.Metadata.RequesterMetadata[MetadataKeyTenantID]
		requestID := p.Metadata.RequesterMetadata[MetadataKeyRequestID]
		if tenantID == "" || requestID == "" {
			s.logger.Printf("cost_events: skipping payload with missing galileo metadata (litellm_id=%s)", p.ID)
			continue
		}
		if err := s.insertCostEvent(r.Context(), p, tenantID, requestID); err != nil {
			s.logger.Printf("cost_events: insert failed for request_id=%s: %v", requestID, err)
			writeErr(w, http.StatusInternalServerError, "insert failed")
			return
		}
		written++
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = writeJSON(w, map[string]any{"written": written})
}

// checkCostEventsAuth uses constant-time comparison so the shared
// secret check doesn't leak length or prefix information through
// timing. The expected secret is configured via the
// GALILEO_COST_EVENTS_SECRET env var read in main.go and threaded onto
// the Server struct.
func (s *Server) checkCostEventsAuth(r *http.Request) bool {
	if s.costEventsSecret == "" {
		return false
	}
	hdr := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(hdr, prefix) {
		return false
	}
	got := strings.TrimPrefix(hdr, prefix)
	return subtle.ConstantTimeCompare([]byte(got), []byte(s.costEventsSecret)) == 1
}

func (s *Server) insertCostEvent(ctx context.Context, p liteLLMPayload, tenantID, requestID string) error {
	costCents := dollarsToCents(p.ResponseCost)
	ts := time.Unix(int64(p.EndTime), int64((p.EndTime-math.Floor(p.EndTime))*1e9)).UTC()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO public.cost_events
			(request_id, tenant_id, event_ts, cost_cents, provider, model, litellm_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (request_id) DO NOTHING
	`, requestID, tenantID, ts, costCents, p.CustomLLMProvider, p.Model, p.ID)
	if err != nil {
		return fmt.Errorf("insert cost_events: %w", err)
	}
	return nil
}

// dollarsToCents converts LiteLLM's float-dollar response_cost to the
// BIGINT cents we store. Half-up rounding so a $0.005 cost becomes 1
// cent rather than silently truncating to 0. Negative costs (shouldn't
// happen) are clamped to 0 — schema CHECK would reject otherwise.
func dollarsToCents(d float64) int64 {
	if d < 0 {
		return 0
	}
	return int64(math.Round(d * 100))
}
