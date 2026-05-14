// galileo-gateway HTTP server. Stage 0 capability matrix:
//
// PR-A:
//   - JWT verification on every request (Ed25519, see kernel/auth)
//   - TenantContext resolution: Postgres-fresh read of monthly_budget_cents
//     on every request (no JWT-cached fallback; Drift-1)
//   - LiteLLM passthrough — proxies the request body to the LiteLLM
//     container
//
// PR-B:
//   - Budget cap enforcement (sum cost_events vs monthly_budget_cents,
//     HTTP 402 deny on overage)
//   - cost_events ingestion via /internal/cost-events webhook (LiteLLM
//     generic_api callback POSTs StandardLoggingPayload here)
//   - OpenTelemetry span emission per request (Jaeger + OTel collector;
//     second plan-deviation, see docs/decisions/0004-observability-substrate.md)
//   - galileo_request_id generation + body metadata injection (so the
//     callback can correlate spend back to the gateway-issued request)
//
// Out of scope for PR-B (lands in PR-C):
//   - workflow ID propagation via AgentOutput.cost_event_request_ids
//     (Drift-2; the proto field exists in PR-B but no Go consumer yet)
//   - agent-runner, Hello Agent, web UI

package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Server wires HTTP routes, auth middleware, tenant resolver, budget
// middleware, OTel middleware, the LiteLLM forwarder, and the
// cost_events webhook receiver.
type Server struct {
	addr             string
	tenants          *TenantResolver
	llm              *LiteLLMClient
	pool             *pgxpool.Pool
	pubKeyPath       string
	logger           *log.Logger
	costEventsSecret string
}

// NewServer constructs a Server. Callers wire dependencies; ListenAndServe
// runs the HTTP loop until ctx cancels or http.Server.Shutdown is called.
func NewServer(addr, pubKeyPath, costEventsSecret string, tenants *TenantResolver, llm *LiteLLMClient, pool *pgxpool.Pool, logger *log.Logger) *Server {
	return &Server{
		addr:             addr,
		tenants:          tenants,
		llm:              llm,
		pool:             pool,
		pubKeyPath:       pubKeyPath,
		logger:           logger,
		costEventsSecret: costEventsSecret,
	}
}

// Handler returns the http.Handler with all routes mounted. Exposed so
// the integration test can hit the same handler without binding a port.
//
// Middleware order on the chat-completions path:
//   tracingMiddleware (root span)
//     → authMiddleware (JWT verify + Postgres-fresh tenant resolve)
//       → budgetMiddleware (sum cost_events vs cap, 402 on overage)
//         → chatCompletions handler
//
// /internal/cost-events is *not* behind authMiddleware — it auths via a
// shared secret in Authorization: Bearer (set by GALILEO_COST_EVENTS_SECRET
// on both this gateway and LiteLLM's GENERIC_LOGGER_HEADERS env). LiteLLM
// is not a tenant; we don't issue it a JWT.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.Handle("POST /v1/chat/completions",
		tracingMiddleware(s.authMiddleware(s.budgetMiddleware(http.HandlerFunc(s.chatCompletions)))))
	mux.HandleFunc("POST /internal/cost-events", s.costEventsHandler)
	return mux
}

// ListenAndServe runs the HTTP server. Returns the first non-nil error
// from http.Server.ListenAndServe or context cancellation.
func (s *Server) ListenAndServe(ctx context.Context) error {
	srv := &http.Server{
		Addr:              s.addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "ok\n")
}

// authMiddleware verifies the Authorization: Bearer <jwt> header and
// resolves the TenantContext fresh from Postgres on every call.
// Drift-1: no JWT-cached fallback. If Postgres is unreachable, the
// request is denied with 503 rather than served on stale cached values.
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hdr := r.Header.Get("Authorization")
		if !strings.HasPrefix(hdr, "Bearer ") {
			writeErr(w, http.StatusUnauthorized, "missing or malformed Authorization header")
			return
		}
		raw := strings.TrimPrefix(hdr, "Bearer ")
		ctx, err := s.tenants.Resolve(r.Context(), s.pubKeyPath, raw)
		if err != nil {
			s.logger.Printf("auth: tenant resolve failed: %v", err)
			if errors.Is(err, ErrPostgresUnavailable) {
				writeErr(w, http.StatusServiceUnavailable, "tenant store unavailable")
				return
			}
			writeErr(w, http.StatusUnauthorized, "invalid token")
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// chatCompletions reads the request body, generates a galileo_request_id,
// forwards to LiteLLM with the id + tenant_id injected into the body's
// metadata field, and returns the response unchanged.
func (s *Server) chatCompletions(w http.ResponseWriter, r *http.Request) {
	tc := TenantFromContext(r.Context())
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "read body: "+err.Error())
		return
	}
	requestID := uuid.Must(uuid.NewV7()).String()
	SetSpanTenantAttrs(r.Context(), tc.TenantID, requestID)

	resp, err := s.llm.Forward(r.Context(), "/v1/chat/completions", body, tc.TenantID, requestID)
	if err != nil {
		s.logger.Printf("litellm: forward failed for tenant=%s request_id=%s: %v", tc.TenantID, requestID, err)
		writeErr(w, http.StatusBadGateway, "upstream failed")
		return
	}
	defer func() { _ = resp.Body.Close() }()
	w.Header().Set("x-galileo-request-id", requestID)
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// writeJSON encodes v as JSON into w. Content-Type and status must be
// set by the caller before invoking. Errors from the encoder are
// returned for the caller to log; the response has already begun
// streaming by the time encoding starts, so there's no useful
// HTTP-level recovery.
func writeJSON(w http.ResponseWriter, v any) error {
	return json.NewEncoder(w).Encode(v)
}
