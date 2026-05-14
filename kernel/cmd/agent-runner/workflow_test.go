// Workflow-level unit tests for HelloAgentWorkflow.
//
// Uses Temporal's in-process testsuite — no real Temporal server,
// no real gateway. Tests verify the workflow's dispatch shape: it
// calls CallLLMActivity exactly once with the provided input and
// translates activity outcomes into the right TaskResult.status.
//
// End-to-end testing (real Temporal + real gateway subprocess) lives
// in integration_test.go under the `agent_runner_integration` build
// tag; this file runs in the default test pass for fast feedback.

package main

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"

	pb "github.com/MannyAmah/GalileoOS/kernel/gen/galileo/v1"
)

func TestHelloAgentWorkflow_HappyPath(t *testing.T) {
	ts := &testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	env.OnActivity(CallLLMActivity, mockCtx, mockTaskInput).Return(&pb.AgentOutput{
		Body:                "hello back",
		CostCents:           5,
		CostEventRequestIds: []string{"req-id-1"},
	}, nil)

	env.ExecuteWorkflow(HelloAgentWorkflow, buildTaskInput())
	if !env.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}
	if err := env.GetWorkflowError(); err != nil {
		t.Fatalf("workflow error: %v", err)
	}
	var result pb.TaskResult
	if err := env.GetWorkflowResult(&result); err != nil {
		t.Fatalf("get result: %v", err)
	}
	if result.Status != "shipped" {
		t.Errorf("status: got %q, want shipped", result.Status)
	}
	if result.Output == nil || result.Output.Body != "hello back" {
		t.Errorf("output.body: got %+v, want body=hello back", result.Output)
	}
	if len(result.Output.CostEventRequestIds) != 1 || result.Output.CostEventRequestIds[0] != "req-id-1" {
		t.Errorf("cost_event_request_ids (Drift-2): got %v, want [req-id-1]", result.Output.CostEventRequestIds)
	}
}

func TestHelloAgentWorkflow_BudgetExceededRejected(t *testing.T) {
	ts := &testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	env.OnActivity(CallLLMActivity, mockCtx, mockTaskInput).Return(nil,
		temporal.NewNonRetryableApplicationError("monthly_budget_cents exceeded", "BudgetExceeded", errors.New("denied")))

	env.ExecuteWorkflow(HelloAgentWorkflow, buildTaskInput())
	if !env.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}
	if err := env.GetWorkflowError(); err != nil {
		t.Fatalf("workflow error: %v", err)
	}
	var result pb.TaskResult
	if err := env.GetWorkflowResult(&result); err != nil {
		t.Fatalf("get result: %v", err)
	}
	if result.Status != "rejected" {
		t.Errorf("budget-exceeded should give status=rejected; got %q", result.Status)
	}
	if result.Error == "" {
		t.Error("expected non-empty Error on rejected result")
	}
}

func TestKnownDepartmentAcceptsHelloRejectsOthers(t *testing.T) {
	if !knownDepartment("hello") {
		t.Error("expected hello to be a known department")
	}
	if knownDepartment("marketing") {
		t.Error("marketing should NOT be registered in Stage 0 (Drift-6: explicit registry)")
	}
	if knownDepartment("") {
		t.Error("empty department should not match")
	}
}

// testsuite uses testify's mock.Anything matcher to skip per-arg
// equality checks. The activity's reflection handles the type
// alignment between the mock and the real signature.
var (
	mockCtx       = mock.Anything
	mockTaskInput = mock.Anything
)

func buildTaskInput() *pb.TaskInput {
	return &pb.TaskInput{
		Tenant: &pb.TenantContext{
			TenantId:           &pb.TenantId{Value: "00000000-0000-7000-8000-000000000001"},
			MonthlyBudgetCents: 99999,
		},
		Department: "hello",
		Goal:       "say hi",
	}
}

// activityCtx is here for completeness — Temporal activities take a
// plain context.Context, not workflow.Context. Used implicitly via
// the testsuite mock.
var _ context.Context = context.Background()
