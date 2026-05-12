# `mcp-servers/` — Custom Galileo MCP servers + vendored fallback servers

Empty in Stage 0. Stage 0 Week 2 vendors a pinned commit of `strukto-ai/mirage` under `mirage-vendored/` for the probe. If the probe fails, the discrete-MCP fallback set from STAGE_0_PLAN.md (`@modelcontextprotocol/server-{github,slack,gdrive}`) is vendored under `fallback/`.

Stage 1 adds the three custom Galileo servers (`galileo-brain-mcp`, `galileo-org-mcp`, `galileo-skill-mcp`) in Go.

See [`docs/galileo_os_infrastructure_plan.md`](../docs/galileo_os_infrastructure_plan.md) §4.6 / Layer 5.
