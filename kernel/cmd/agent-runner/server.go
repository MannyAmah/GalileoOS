// HTTP server for the agent-runner.
//
// Internal control-plane API (web → agent-runner). Two endpoints:
//
//   POST /v1/tasks
//     Body: TaskInput JSON (snake_case keys per CLAUDE.md "HTTP types
//     as JSON schema"). Auth: tenant JWT in Authorization: Bearer
//     (same Ed25519 keypair the gateway verifies).
//     Action: Validates department against the registry, mints a
//     Temporal workflow ID as `agent-task-<uuidv7>`, calls
//     client.ExecuteWorkflow with the registered workflow Name (the
//     department slug), returns {task_id: <workflowID>} for polling.
//
//   GET /v1/tasks/{id}
//     Auth: tenant JWT.
//     Action: Polls the Temporal workflow's status. Returns
//     {status: "running"} if the workflow is still executing,
//     or the TaskResult JSON when complete. The web polls this every
//     500ms.
//
// JWT verification uses the same kernel/auth public key the gateway
// loads (default kernel/auth/dev-keys/public.pem). The agent-runner
// accepts tenant JWTs minted by the same issuer; it does not have a
// service-account identity of its own in Stage 0.

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
	"go.temporal.io/sdk/client"

	"github.com/MannyAmah/GalileoOS/kernel/auth"
	pb "github.com/MannyAmah/GalileoOS/kernel/gen/galileo/v1"
)

// Server holds the HTTP routes + the Temporal client used to start
// and poll workflows. Caller owns the temporal client's lifecycle.
type Server struct {
	addr       string
	pubKeyPath string
	temporal   client.Client
	logger     *log.Logger
}

func NewServer(addr, pubKeyPath string, c client.Client, logger *log.Logger) *Server {
	return &Server{addr: addr, pubKeyPath: pubKeyPath, temporal: c, logger: logger}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.Handle("POST /v1/tasks", s.authMiddleware(http.HandlerFunc(s.createTask)))
	mux.Handle("GET /v1/tasks/{id}", s.authMiddleware(http.HandlerFunc(s.getTask)))
	return mux
}

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

// authMiddleware verifies the tenant JWT against the same public key
// the gateway uses. The middleware doesn't construct a TenantContext —
// the workflow's input carries it from the request body, so we don't
// duplicate gateway's Postgres-fresh lookup here. Stage 0: trust the
// body's tenant_id matches the JWT's tenant_id (assertion below).
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hdr := r.Header.Get("Authorization")
		if !strings.HasPrefix(hdr, "Bearer ") {
			writeErr(w, http.StatusUnauthorized, "missing or malformed Authorization header")
			return
		}
		raw := strings.TrimPrefix(hdr, "Bearer ")
		claims, err := auth.VerifyToken(s.pubKeyPath, raw)
		if err != nil {
			writeErr(w, http.StatusUnauthorized, "invalid token")
			return
		}
		ctx := context.WithValue(r.Context(), authedTenantKey{}, claims.TenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type authedTenantKey struct{}

func authedTenant(ctx context.Context) string {
	v, _ := ctx.Value(authedTenantKey{}).(string)
	return v
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "read body: "+err.Error())
		return
	}
	var input pb.TaskInput
	// Use standard encoding/json with snake_case (matches the json:
	// tags on the generated structs). See CLAUDE.md "HTTP types as
	// JSON schema" for why we don't use protojson here.
	if err := json.Unmarshal(body, &input); err != nil {
		writeErr(w, http.StatusBadRequest, "parse TaskInput: "+err.Error())
		return
	}
	if !knownDepartment(input.GetDepartment()) {
		writeErr(w, http.StatusBadRequest, "unknown department: "+input.GetDepartment())
		return
	}
	// Stage 0 assertion: the JWT's tenant_id must match the body's
	// tenant_id. Prevents a tenant from spoofing tasks as another
	// tenant via a valid token of their own.
	if input.GetTenant().GetTenantId().GetValue() != authedTenant(r.Context()) {
		writeErr(w, http.StatusForbidden, "tenant_id mismatch between JWT and body")
		return
	}

	workflowID := "agent-task-" + uuid.Must(uuid.NewV7()).String()
	_, err = s.temporal.ExecuteWorkflow(r.Context(),
		client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: TaskQueue,
		},
		input.GetDepartment(), // workflow name from the registry
		&input,
	)
	if err != nil {
		s.logger.Printf("ExecuteWorkflow failed for tenant=%s: %v", authedTenant(r.Context()), err)
		writeErr(w, http.StatusServiceUnavailable, "could not start workflow")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{"task_id": workflowID})
}

func (s *Server) getTask(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("id")
	if taskID == "" {
		writeErr(w, http.StatusBadRequest, "missing task id")
		return
	}
	// Describe the workflow to check whether it's done. If not yet
	// complete, return {status: "running"} so the web's poll loop
	// continues. If complete, fetch the result and return TaskResult.
	descCtx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	desc, err := s.temporal.DescribeWorkflowExecution(descCtx, taskID, "")
	if err != nil {
		writeErr(w, http.StatusNotFound, "task not found")
		return
	}
	info := desc.GetWorkflowExecutionInfo()
	if info == nil || info.GetStatus() == 1 /* WORKFLOW_EXECUTION_STATUS_RUNNING */ {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "running"})
		return
	}

	getCtx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	we := s.temporal.GetWorkflow(getCtx, taskID, "")
	var result pb.TaskResult
	if err := we.Get(getCtx, &result); err != nil {
		writeErr(w, http.StatusInternalServerError, "fetch workflow result: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	// Standard encoding/json reads the json:"..." tags emitted by
	// protoc-gen-go (snake_case), matching the body-side decode in
	// createTask. See CLAUDE.md "HTTP types as JSON schema".
	_ = json.NewEncoder(w).Encode(&result)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
