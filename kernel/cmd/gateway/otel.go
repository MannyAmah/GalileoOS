// OpenTelemetry tracer wiring (Stage 0 observability substrate).
//
// Substrate decision: Jaeger + OTel collector substitute for the
// originally-named Opik. See docs/closeouts/CLOSEOUT_OPIK_RECONSIDERED.md
// and docs/decisions/0004-observability-substrate.md for the second
// plan-deviation in the project's history. Trigger: the dependency-shape
// discovery pass surfaced that Opik's self-hosted deployment is a
// 13-container platform, not a single observability service.
//
// Span shape, Stage 0:
//
//   gateway.request                       (root span per HTTP request)
//     ├─ gateway.tenant.resolve           (TenantResolver.Resolve)
//     └─ gateway.litellm.forward          (LiteLLMClient.Forward)
//
// Attributes set on the root span:
//   tenant_id, model (when known from request), galileo_request_id,
//   cost_cents (set asynchronously by the cost_events webhook;
//   correlation is via galileo_request_id, not parent/child).

package main

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// SpanAttrTenantID names the OTel attribute key holding the tenant UUID.
// Used in dashboard queries to filter spans per tenant.
const SpanAttrTenantID = "galileo.tenant_id"

// SpanAttrRequestID is the gateway-generated UUIDv7. Joined against
// cost_events.request_id for cost-vs-latency dashboards.
const SpanAttrRequestID = "galileo.request_id"

// InitTracer configures a TracerProvider that exports to the OTel
// collector via OTLP gRPC. endpoint is the host:port of the collector
// (otel-collector:4317 in compose, localhost:4317 from the host or CI).
// Returns a shutdown closure that flushes pending spans on process exit.
func InitTracer(ctx context.Context, endpoint string) (func(context.Context) error, error) {
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("otel exporter: %w", err)
	}
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("otel resource: %w", err)
	}
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	return tp.Shutdown, nil
}

// tracingMiddleware wraps the handler chain in a root span per request.
// Attributes added later (tenant_id, request_id, model) by the
// downstream handlers via SetAttributes on the same span.
func tracingMiddleware(next http.Handler) http.Handler {
	tracer := otel.Tracer(serviceName)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "gateway.request",
			oteltrace.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.route", r.URL.Path),
			),
		)
		defer span.End()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SetSpanTenantAttrs attaches galileo-specific attributes to the
// current span. Called from the auth middleware once the tenant is
// resolved and from the chat handler once the request body is parsed.
func SetSpanTenantAttrs(ctx context.Context, tenantID, requestID string) {
	span := oteltrace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}
	span.SetAttributes(
		attribute.String(SpanAttrTenantID, tenantID),
		attribute.String(SpanAttrRequestID, requestID),
	)
}
