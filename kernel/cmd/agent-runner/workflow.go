// HelloAgentWorkflow — Stage 0's demo agent.
//
// One workflow, one activity, one LLM call. The simplest possible
// agent: takes a goal, asks the model, returns the model's reply
// alongside the cost_events.request_id the gateway generated.
//
// Stage 0's 100×-runs gate test exercises this exact path: web POST →
// agent-runner HTTP → ExecuteWorkflow → ExecuteActivity → gateway →
// LiteLLM → cost_events row written → workflow returns AgentOutput
// → web polls + renders.
//
// The workflow body is intentionally trivial. The real work is the
// activity (which talks to the outside world). Workflows must stay
// deterministic — no clocks, no random, no network — and a thin
// dispatch layer is the Temporal canonical shape.

package main

import (
	"time"

	"go.temporal.io/sdk/workflow"

	pb "github.com/MannyAmah/GalileoOS/kernel/gen/galileo/v1"
)

// TaskQueue is the Temporal task queue the runner subscribes to and
// the HTTP server uses when starting workflows. Stage 0 uses one
// queue for all departments; Stage 1's department fan-out may shard.
const TaskQueue = "galileo-agent-runner"

// HelloAgentWorkflow executes one CallLLMActivity and returns the
// resulting AgentOutput as TaskResult.output with status "shipped".
// Failure modes (activity error, budget cap deny) surface as status
// "rejected" with the error string set.
func HelloAgentWorkflow(ctx workflow.Context, input *pb.TaskInput) (*pb.TaskResult, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 90 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("HelloAgentWorkflow started", "department", input.GetDepartment(), "goal_len", len(input.GetGoal()))

	var output pb.AgentOutput
	if err := workflow.ExecuteActivity(ctx, CallLLMActivity, input).Get(ctx, &output); err != nil {
		logger.Error("CallLLMActivity failed", "err", err)
		return &pb.TaskResult{
			Status: "rejected",
			Error:  err.Error(),
		}, nil
	}
	return &pb.TaskResult{
		Status: "shipped",
		Output: &output,
	}, nil
}
