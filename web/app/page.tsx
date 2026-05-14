// Server component shell for the Hello Agent demo. The interactive UI
// lives in app/hello-agent.tsx ("use client"); this file stays a
// server component so the static page chrome doesn't get bundled with
// React's client runtime unnecessarily. Server/client boundary is the
// import below.

import HelloAgent from "./hello-agent";

export default function Page() {
  return (
    <main style={{ padding: 24, fontFamily: "system-ui, sans-serif", maxWidth: 720 }}>
      <h1>Galileo OS — Hello Agent</h1>
      <p style={{ color: "#555" }}>
        Stage 0 demo. Submits a goal to the agent-runner, which runs a Temporal
        workflow that calls the gateway, which forwards to LiteLLM. The result
        comes back with the gateway-issued <code>cost_events.request_id</code>
        (Drift-2 correlation).
      </p>
      <HelloAgent />
    </main>
  );
}
