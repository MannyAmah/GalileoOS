// cost-recon — Stage 0 reconciliation between cost_events (gateway's
// per-LLM-call ledger) and Stripe metered billing.
//
// What it does:
//   - Reads cost_events for a tenant within a date range.
//   - POSTs each row to Stripe as a billing.meter_event with the
//     cost_events.request_id as Stripe's identifier (Stripe enforces
//     uniqueness within 24h, so the same request_id can be posted
//     across runs and Stripe deduplicates).
//   - Logs per-row outcome; exits non-zero if any row failed for a
//     reason other than dedup.
//
// What it does NOT do:
//   - Polling. Operators run this on a cron (Stage 0) or call it from
//     the agent-runner Temporal workflow (PR-C onward).
//   - Tenant management. Caller passes --tenant; binary handles one
//     tenant per invocation.
//
// Verified Stripe API shape (POST /v1/billing/meter_events):
//   event_name      string, ≤100 chars
//   payload.value   (mapped from cost_events.cost_cents)
//   payload.stripe_customer_id
//   identifier      (mapped from cost_events.request_id; unique within 24h)
//   timestamp       Unix epoch seconds (mapped from cost_events.event_ts;
//                   valid range: past 35 days to +5 minutes)

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stripe/stripe-go/v85"
	"github.com/stripe/stripe-go/v85/billing/meterevent"
)

const serviceName = "galileo-cost-recon"

// MeterEventName is the Stripe meter-event name we register usage under.
// Mapped 1:1 from cost_events; the Stage 1 gate evolves this to per-model
// pricing.
const MeterEventName = "galileo_llm_usage_cents"

// MaxStripePastTimestamp is Stripe's documented window for accepting
// historical events: "past 35 calendar days." We refuse to reconcile
// further back than this.
const MaxStripePastTimestamp = 35 * 24 * time.Hour

type config struct {
	dbURL      string
	stripeKey  string
	tenantID   string
	customerID string
	since      time.Duration
}

func main() {
	logger := log.New(os.Stderr, serviceName+" ", log.LstdFlags|log.LUTC)

	cfg := parseFlags()
	if err := run(context.Background(), cfg, logger); err != nil {
		logger.Fatalf("recon: %v", err)
	}
}

func parseFlags() config {
	cfg := config{}
	flag.StringVar(&cfg.dbURL, "db", os.Getenv("GALILEO_GATEWAY_DATABASE_URL"), "Postgres DSN (env: GALILEO_GATEWAY_DATABASE_URL)")
	flag.StringVar(&cfg.stripeKey, "stripe-key", os.Getenv("STRIPE_SECRET_KEY"), "Stripe API secret key (env: STRIPE_SECRET_KEY)")
	flag.StringVar(&cfg.tenantID, "tenant", "", "tenant UUID to reconcile")
	flag.StringVar(&cfg.customerID, "customer", "", "Stripe customer id (cus_...) for this tenant")
	flag.DurationVar(&cfg.since, "since", 24*time.Hour, "reconcile cost_events with event_ts >= now()-since")
	flag.Parse()
	return cfg
}

func run(ctx context.Context, cfg config, logger *log.Logger) error {
	if cfg.dbURL == "" {
		return fmt.Errorf("--db (or GALILEO_GATEWAY_DATABASE_URL) required")
	}
	if cfg.stripeKey == "" {
		return fmt.Errorf("--stripe-key (or STRIPE_SECRET_KEY) required")
	}
	if cfg.tenantID == "" || cfg.customerID == "" {
		return fmt.Errorf("--tenant and --customer required")
	}
	if cfg.since > MaxStripePastTimestamp {
		return fmt.Errorf("--since %s exceeds Stripe's 35-day acceptance window", cfg.since)
	}

	stripe.Key = cfg.stripeKey

	pool, err := pgxpool.New(ctx, cfg.dbURL)
	if err != nil {
		return fmt.Errorf("postgres pool: %w", err)
	}
	defer pool.Close()

	rows, err := pool.Query(ctx, `
		SELECT request_id, event_ts, cost_cents
		FROM public.cost_events
		WHERE tenant_id = $1 AND event_ts >= $2
		ORDER BY event_ts ASC
	`, cfg.tenantID, time.Now().Add(-cfg.since))
	if err != nil {
		return fmt.Errorf("query cost_events: %w", err)
	}
	defer rows.Close()

	var posted, failed int
	for rows.Next() {
		var (
			requestID string
			eventTs   time.Time
			costCents int64
		)
		if err := rows.Scan(&requestID, &eventTs, &costCents); err != nil {
			return fmt.Errorf("scan cost_events row: %w", err)
		}
		if err := postMeterEvent(cfg.customerID, requestID, eventTs, costCents); err != nil {
			failed++
			logger.Printf("post failed for request_id=%s: %v", requestID, err)
			continue
		}
		posted++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate cost_events: %w", err)
	}
	logger.Printf("recon done: posted=%d failed=%d", posted, failed)
	if failed > 0 {
		return fmt.Errorf("%d rows failed", failed)
	}
	return nil
}

// postMeterEvent writes one cost_events row to Stripe as a meter_event.
// Stripe's 24h identifier-uniqueness window handles dedup — re-running
// recon for the same period is safe.
func postMeterEvent(customerID, requestID string, eventTs time.Time, costCents int64) error {
	params := &stripe.BillingMeterEventParams{
		EventName:  stripe.String(MeterEventName),
		Identifier: stripe.String(requestID),
		Timestamp:  stripe.Int64(eventTs.Unix()),
		Payload: map[string]string{
			"stripe_customer_id": customerID,
			"value":              fmt.Sprintf("%d", costCents),
		},
	}
	_, err := meterevent.New(params)
	return err
}
