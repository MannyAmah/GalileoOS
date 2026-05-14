# ADR-0004 — Observability substrate: Jaeger + OTel collector instead of Opik

| Field | Value |
| --- | --- |
| **Status** | Accepted |
| **Date** | 2026-05-14 |
| **Decider** | Emmanuel (founder) |
| **Author** | Claude Opus 4.7 (1M context) under Emmanuel's direction |
| **Supersedes** | The Opik observability substrate named in [`docs/galileo_os_infrastructure_plan.md`](../galileo_os_infrastructure_plan.md) §4.5 (Layer 5) and [`docs/plans/STAGE_0_PLAN.md`](../plans/STAGE_0_PLAN.md) §Week 3 deliverable 7 (revised to record the substitution) |
| **Plan deviation** | Yes — **second plan-deviation in the project's history.** Follows the four-part template established by ADR-0003 (closeout + canonical plan edits + this ADR + the code that implements the substitution). |
| **Companion artifacts** | [`docs/closeouts/CLOSEOUT_OPIK_RECONSIDERED.md`](../closeouts/CLOSEOUT_OPIK_RECONSIDERED.md) — full structural finding. [`docs/solutions/SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md`](../solutions/SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md) — the dependency-shape pattern this finding instantiates (2nd of 3 documented instances). |

## Context

The original plan named Opik as the observability backend at Layer 5. During PR-B's inline planning round (2026-05-14), reading Opik's self-hosted deployment documentation directly surfaced the structural mismatch: **Opik's self-hosted deployment is a 13-container platform** (MySQL + Redis + ClickHouse + Zookeeper + MinIO + 5 Opik-specific services + Jaeger + OTel collector). The plan's framing of Opik as a single observability service was incorrect — same shape of finding as PR #13's Mirage SDK-only discovery, and the third instance of the documented "marketing description vs deployment shape" pattern.

Three readings were considered (full reasoning in the companion closeout):

1. **Opik in a stripped-down configuration.** Rejected — no documented minimal-mode deployment; the 13-container set is the supported topology. Operating an unsupported subset is a maintenance burden Galileo doesn't take on for a non-core dependency.
2. **Defer observability past Stage 0.** Rejected — STAGE_0_PLAN.md §Week 3 deliverable 7 names span emission as a Stage 0 gate-test requirement. Relaxing a pre-registered gate is what v7 rule 9 specifically refuses.
3. **Substitute Jaeger + OTel collector.** **Chosen.** Two containers, OTel-native, covers the trace ingest + storage + query surface the gate test requires.

## Decision

The observability substrate at Layer 5 is **Jaeger + OTel collector**:

- `jaegertracing/all-in-one:2.18.0` — trace storage + query + UI in one container, alpine-based.
- `otel/opentelemetry-collector-contrib:0.152.0` — receiver/exporter/extensions; `FROM scratch` distroless image.
- Gateway emits OTLP gRPC to the collector at `:4317`; collector forwards to Jaeger at `:4317`.
- Health check on Jaeger via port 16686 UI; health check on collector via the `health_check` extension on port 13133. Both are polled by the integration test setup with bounded retries (no docker healthcheck because OTel collector's FROM-scratch image has no shell).
- Pin policy: tracks latest stable directly under CLAUDE.md "Service image pins."

## Reversal triggers

This decision should be reconsidered if any of the following fire in Stage 1+:

- An observability stack that Galileo's operators actually want to run (per tenant evaluation feedback) is non-OTel-native and integrates more cleanly with Opik than with Jaeger.
- Gate-test trace volume or query complexity exceeds Jaeger's documented limits at the substrate level (note: Jaeger v2 backs onto pluggable storage including ClickHouse, Elasticsearch, etc., so this trigger is well past Stage 0 scope).
- A Galileo customer requires the LLM-evaluation surface (prompt diff playgrounds, eval dashboards) Opik provides natively that Jaeger doesn't. This would be a Stage 2 product feature, not a Stage 0 infrastructure choice.

Until then, two containers handle Stage 0's observability gate. Reconsider only if a structural trigger fires, not on threshold-based metrics.

## Consequences

**Operational.** Two new compose services (jaeger, otel-collector), two new CI service containers, one CI-only wrapper image (`deploy/compose/otel-wrapper/`) because GHA service-containers can't override CMD or mount config files. Wrapper-vs-upstream divergence is one line of CMD; documented in the wrapper directory's README.

**Code.** Gateway gains an OTel tracer provider (`kernel/cmd/gateway/otel.go`), tracing middleware in the request chain, span attribute attachment for tenant_id + galileo_request_id. Stage 0 span shape: one root span per HTTP request, child spans for tenant resolve + LiteLLM forward.

**Plan & spec.** STAGE_0_PLAN.md §Week 3 deliverable 7 records the substitution; the gate-test requirement is preserved unchanged (100 demo runs → 100 parent spans in the backend).

**Cross-pollination of the dependency-shape pattern.** This finding feeds back into [`docs/solutions/SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md`](../solutions/SOLUTION_DEPENDENCY_SHAPE_VERIFICATION.md) as instance 2; the solutions doc is now load-bearing for future planning rounds' dependency-shape verification step.
