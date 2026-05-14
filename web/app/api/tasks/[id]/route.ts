// GET /api/tasks/[id] — proxy to the agent-runner's task-poll endpoint.
//
// Web polls this every 500ms with an AbortController; route handler
// forwards each poll to the agent-runner. Same auth + URL convention
// as the create-task route.

import type { NextRequest } from "next/server";

const AGENT_URL = process.env.GALILEO_AGENT_RUNNER_URL ?? "http://localhost:8081";

export async function GET(
  request: NextRequest,
  context: { params: Promise<{ id: string }> },
) {
  const { id } = await context.params;
  const upstream = await fetch(`${AGENT_URL}/v1/tasks/${encodeURIComponent(id)}`, {
    method: "GET",
    headers: {
      Authorization: request.headers.get("Authorization") ?? "",
    },
  });
  return new Response(await upstream.text(), {
    status: upstream.status,
    headers: { "Content-Type": "application/json" },
  });
}
