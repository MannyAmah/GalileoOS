// galileo-gateway HTTP server. Stage 0 / PR-A capability:
//
//   - JWT verification on every request (Ed25519, see kernel/auth)
//   - TenantContext resolution: Postgres-fresh read of monthly_budget_cents
//     on every request (no JWT-cached fallback; Drift-1 resolution from
//     Week 3 inline-plan round)
//   - LiteLLM passthrough — proxies the request body to the LiteLLM
//     container, returns the response unchanged
//
// Out of scope for PR-A (lands in PR-B):
//   - cost_events writing on LiteLLM usage callbacks
//   - budget cap enforcement (sum cost_events vs monthly_budget_cents)
//   - observability span emission
//
// Out of scope for PR-A (lands in PR-C):
//   - workflow ID propagation to cost_events rows

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
)

// Server wires HTTP routes, auth middleware, tenant resolver, and the
// LiteLLM passthrough handler.
type Server struct {
	addr       string
	tenants    *TenantResolver
	llm        *LiteLLMClient
	pubKeyPath string
	logger     *log.Logger
}

// NewServer constructs a Server. Callers wire dependencies; ListenAndServe
// runs the HTTP loop until ctx cancels or http.Server.Shutdown is called.
func NewServer(addr, pubKeyPath string, tenants *TenantResolver, llm *LiteLLMClient, logger *log.Logger) *Server {
	return &Server{
		addr:       addr,
		tenants:    tenants,
		llm:        llm,
		pubKeyPath: pubKeyPath,
		logger:     logger,
	}
}

// Handler returns the http.Handler with all routes mounted. Exposed so
// the integration test can hit the same handler without binding a port.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.Handle("POST /v1/chat/completions", s.authMiddleware(http.HandlerFunc(s.chatCompletions)))
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
			// 503 specifically when Postgres is unreachable; 401 for
			// token-level failures.
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

// chatCompletions proxies the request body to LiteLLM and returns the
// response unchanged. PR-B will add usage-event capture and cost_events
// row writing here.
func (s *Server) chatCompletions(w http.ResponseWriter, r *http.Request) {
	tc := TenantFromContext(r.Context())
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "read body: "+err.Error())
		return
	}
	resp, err := s.llm.Forward(r.Context(), "/v1/chat/completions", body)
	if err != nil {
		s.logger.Printf("litellm: forward failed for tenant=%s: %v", tc.TenantID, err)
		writeErr(w, http.StatusBadGateway, "upstream failed")
		return
	}
	defer resp.Body.Close()
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
