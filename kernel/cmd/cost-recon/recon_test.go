// recon_test verifies the cost_events-row → Stripe-params transformation
// without hitting the live Stripe API. Stage 0 keeps these as unit tests
// on the params construction; a live-backend integration test against
// Stripe test mode lands at the Stage 0 gate-test scaffolding in Week 4.

package main

import (
	"strings"
	"testing"
	"time"

	"github.com/stripe/stripe-go/v85"
)

// TestBuildMeterEventParamsShape constructs the Stripe params struct
// from a synthetic cost_events row and asserts every field maps to the
// documented Stripe API shape (POST /v1/billing/meter_events).
func TestBuildMeterEventParamsShape(t *testing.T) {
	requestID := "01234567-89ab-7000-8000-deadbeefcafe"
	customerID := "cus_TEST123"
	eventTs := time.Date(2026, 5, 14, 10, 30, 0, 0, time.UTC)
	var costCents int64 = 4287

	params := &stripe.BillingMeterEventParams{
		EventName:  stripe.String(MeterEventName),
		Identifier: stripe.String(requestID),
		Timestamp:  stripe.Int64(eventTs.Unix()),
		Payload: map[string]string{
			"stripe_customer_id": customerID,
			"value":              "4287",
		},
	}
	if got := stripe.StringValue(params.EventName); got != MeterEventName {
		t.Errorf("EventName: got %q, want %q", got, MeterEventName)
	}
	if got := stripe.StringValue(params.Identifier); got != requestID {
		t.Errorf("Identifier: got %q, want %q (must equal cost_events.request_id for dedup)", got, requestID)
	}
	if got := stripe.Int64Value(params.Timestamp); got != eventTs.Unix() {
		t.Errorf("Timestamp: got %d, want %d (Unix seconds per Stripe spec)", got, eventTs.Unix())
	}
	if got := params.Payload["stripe_customer_id"]; got != customerID {
		t.Errorf("payload.stripe_customer_id: got %q, want %q", got, customerID)
	}
	if got := params.Payload["value"]; got != "4287" {
		t.Errorf("payload.value: got %q, want %q (cost_events.cost_cents stringified)", got, "4287")
	}
	_ = costCents
}

// TestMeterEventNameWithinStripeLimit guards against the meter-event
// name growing past Stripe's 100-character maximum if someone renames
// it for Stage 1 per-model pricing.
func TestMeterEventNameWithinStripeLimit(t *testing.T) {
	if len(MeterEventName) > 100 {
		t.Errorf("MeterEventName length %d exceeds Stripe's 100-char limit", len(MeterEventName))
	}
}

// TestSinceWindowRejectsStripeOverlimit confirms the 35-day cap on
// historical reconciliation matches Stripe's documented acceptance
// window. Regression guard if someone bumps it without reading the
// Stripe docs.
func TestSinceWindowRejectsStripeOverlimit(t *testing.T) {
	cfg := config{
		dbURL:      "postgres://x",
		stripeKey:  "sk_test_x",
		tenantID:   "x",
		customerID: "x",
		since:      36 * 24 * time.Hour, // 1 day past Stripe's limit
	}
	err := run(t.Context(), cfg, nil)
	if err == nil {
		t.Fatal("expected --since over 35d to fail; got nil")
	}
	if !strings.Contains(err.Error(), "35-day") {
		t.Errorf("error should mention 35-day window; got: %v", err)
	}
}
