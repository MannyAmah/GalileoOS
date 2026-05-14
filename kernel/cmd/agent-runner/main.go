// Package main is the Galileo agent runner entry point.
//
// Stage 0 / PR-C capability: Temporal worker subscribed to the
// `galileo-agent-runner` task queue, registered workflow per
// the explicit registry (Drift-6), HTTP server on :8081 for
// /v1/tasks{,/{id}}. JWT verification reuses kernel/auth.
//
// Environment variables (* = required):
//
//   GALILEO_AGENT_ADDR                 listen address     default ":8081"
//   GALILEO_AGENT_PUBKEY               JWT public key     default "kernel/auth/dev-keys/public.pem"
//   GALILEO_AGENT_JWT_PRIVATE_KEY_PATH JWT private key    default "kernel/auth/dev-keys/private.pem"
//                                                          (used by the activity to mint tokens
//                                                           when calling the gateway)
//   GALILEO_AGENT_GATEWAY_URL          gateway base URL   default "http://localhost:8080"
//   GALILEO_AGENT_TEMPORAL_HOSTPORT    Temporal frontend  default "localhost:7233"
//   GALILEO_AGENT_TEMPORAL_NAMESPACE   Temporal namespace default "default"
//
// Boot sequence (fail-loud at each step):
//
//   1. Dial Temporal frontend
//   2. Build worker + register workflows + register activity
//   3. Start worker (background)
//   4. HTTP server foreground until ctx cancels
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

const serviceName = "galileo-agent-runner"

func main() {
	logger := log.New(os.Stderr, serviceName+" ", log.LstdFlags|log.LUTC)

	addr := envOr("GALILEO_AGENT_ADDR", ":8081")
	pubKeyPath := envOr("GALILEO_AGENT_PUBKEY", "kernel/auth/dev-keys/public.pem")
	temporalHostPort := envOr("GALILEO_AGENT_TEMPORAL_HOSTPORT", "localhost:7233")
	temporalNamespace := envOr("GALILEO_AGENT_TEMPORAL_NAMESPACE", "default")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	c, err := client.Dial(client.Options{
		HostPort:  temporalHostPort,
		Namespace: temporalNamespace,
	})
	if err != nil {
		logger.Fatalf("temporal dial: %v", err)
	}
	defer c.Close()

	w := worker.New(c, TaskQueue, worker.Options{})
	registerWorkflows(w)
	w.RegisterActivity(CallLLMActivity)

	if err := w.Start(); err != nil {
		logger.Fatalf("worker start: %v", err)
	}
	defer w.Stop()

	srv := NewServer(addr, pubKeyPath, c, logger)
	logger.Printf("listening on %s", addr)
	if err := srv.ListenAndServe(ctx); err != nil {
		logger.Fatalf("http server: %v", err)
	}
}
