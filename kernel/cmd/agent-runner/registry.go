// Explicit workflow registry (Drift-6 resolution).
//
// CLAUDE.md locks in: "explicit workflow registry per Drift-6 — map of
// department slug to workflow function, not convention-based dispatch."
// Stage 0 only ships "hello"; future departments (marketing, sales,
// support, etc.) add entries to this map explicitly, never via
// reflection or name-based resolution.
//
// The registry doubles as the worker-registration source of truth: the
// worker iterates this map at boot and calls RegisterWorkflowWithOptions
// with the slug as the Name. The HTTP server uses the same map to
// validate the `department` field on incoming TaskInput requests.

package main

import (
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

// workflowRegistry maps department slugs to workflow functions. Stage 0
// only has "hello". Stage 1's department fan-out adds entries here.
var workflowRegistry = map[string]any{
	"hello": HelloAgentWorkflow,
}

// registerWorkflows binds every workflow in the registry to w, using
// the slug as the workflow Name (what client.ExecuteWorkflow sends on
// the wire). Activities are registered separately in main.go.
func registerWorkflows(w worker.Worker) {
	for slug, fn := range workflowRegistry {
		w.RegisterWorkflowWithOptions(fn, workflow.RegisterOptions{Name: slug})
	}
}

// knownDepartment reports whether slug is a department this runner
// can dispatch. Used by the HTTP server to reject TaskInputs with
// unknown department slugs before they hit Temporal.
func knownDepartment(slug string) bool {
	_, ok := workflowRegistry[slug]
	return ok
}
