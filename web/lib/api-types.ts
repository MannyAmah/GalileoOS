// Hand-written TypeScript interfaces for the agent-runner wire format.
// Tracked alongside changes to schemas/galileo/v1/*.proto — when a
// proto field changes, this file changes in the same PR. See
// CLAUDE.md "HTTP types as JSON schema" for the convention; reconsider
// generated TypeScript (e.g., bufbuild/es) when web complexity grows
// to 3+ entity types or a second TypeScript client appears.
//
// Field names are snake_case to match what `encoding/json` emits from
// the protoc-gen-go generated structs (it reads the auto-emitted
// `json:"snake_case"` tags). camelCase would require switching the
// Go side to `protojson` — different tradeoff, not Stage 0's choice.

export interface TenantId {
  value: string;
}

export interface TenantContext {
  tenant_id: TenantId;
  plan_tier?: string;
  monthly_budget_cents: number;
  issued_at_unix?: number;
}

export interface TaskInput {
  tenant: TenantContext;
  department: string;
  goal: string;
}

export interface ToolCall {
  tool: string;
  args_json: string;
  destructive: boolean;
}

export interface AgentOutput {
  body: string;
  tool_calls?: ToolCall[];
  cost_cents: number;
  cost_event_request_ids: string[];
}

export interface TaskResult {
  status: "shipped" | "rejected" | "halted_for_approval";
  output?: AgentOutput;
  error?: string;
}

// The agent-runner's create-task response: 202 + this body.
export interface CreateTaskResponse {
  task_id: string;
}

// The agent-runner's poll-task response is either a running marker
// or a TaskResult once the workflow completes. The web differentiates
// by checking for the `status` field.
export interface PollRunning {
  status: "running";
}
export type PollResponse = PollRunning | TaskResult;
