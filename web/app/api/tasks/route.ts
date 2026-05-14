// POST /api/tasks — proxy to the agent-runner.
//
// The browser only talks to the Next.js server; this route handler is
// server-side and forwards the request to agent-runner on
// GALILEO_AGENT_RUNNER_URL (default http://localhost:8081). Avoids CORS
// configuration and keeps the agent-runner port off the public network
// when web is reverse-proxied in production.
//
// Authorization header passes through unchanged — the agent-runner
// verifies the tenant JWT against the same kernel/auth public key the
// gateway uses.

import type { NextRequest } from "next/server";

const AGENT_URL = process.env.GALILEO_AGENT_RUNNER_URL ?? "http://localhost:8081";

export async function POST(request: NextRequest) {
  const body = await request.text();
  const upstream = await fetch(`${AGENT_URL}/v1/tasks`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: request.headers.get("Authorization") ?? "",
    },
    body,
  });
  return new Response(await upstream.text(), {
    status: upstream.status,
    headers: { "Content-Type": "application/json" },
  });
}
