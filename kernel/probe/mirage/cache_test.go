package mirage

import (
	"context"
	"testing"
	"time"
)

// Synthetic test parameters chosen so each test completes in ~200ms.
// Production gate criteria (50 concurrent / 100ms p99) are imposed by
// the caller in plan-PR #11, not the apparatus itself.
const (
	cacheTestConcurrency = 8
	cacheTestDuration    = 200 * time.Millisecond
	cacheTestTTL         = 5 * time.Second
)

func TestRunCacheProbe_HappyPath(t *testing.T) {
	base := newBaseMock()
	base.seedTenant("t0", map[string][]byte{}) // ensure tenant exists
	res, err := RunCacheProbe(context.Background(), base, "t0", cacheTestConcurrency, cacheTestDuration, cacheTestTTL, 50.0 /*ms*/)
	t.Logf("seed=%d p99=%.2fms reads=%d writes=%d", res.Seed, res.P99ReadLatencyMs, res.Reads, res.Writes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Pass {
		t.Fatalf("apparatus reported failure on happy-path mock: %+v", res)
	}
}

func TestRunCacheProbe_DetectsP99Exceeded(t *testing.T) {
	base := newBaseMock()
	base.seedTenant("t0", map[string][]byte{})
	slow := &slowMock{baseMock: base}
	res, err := RunCacheProbe(context.Background(), slow, "t0", cacheTestConcurrency, cacheTestDuration, cacheTestTTL, 5.0 /*ms*/)
	t.Logf("seed=%d p99=%.2fms (slow injection 10ms; threshold 5ms)", res.Seed, res.P99ReadLatencyMs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Pass {
		t.Fatalf("apparatus failed to detect p99 exceeded: %+v", res)
	}
	if res.P99ReadLatencyMs < 5 {
		t.Fatalf("expected p99 >= 5ms with 10ms injection; got %.2fms", res.P99ReadLatencyMs)
	}
}

func TestRunCacheProbe_DetectsCorruption(t *testing.T) {
	base := newBaseMock()
	base.seedTenant("t0", map[string][]byte{})
	corrupting := &corruptingMock{baseMock: base}
	res, err := RunCacheProbe(context.Background(), corrupting, "t0", cacheTestConcurrency, cacheTestDuration, cacheTestTTL, 50.0)
	t.Logf("seed=%d corruption=%d stale=%d", res.Seed, res.CorruptionCount, res.StaleReadCount)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Pass {
		t.Fatalf("apparatus failed to detect corruption: %+v", res)
	}
	if res.CorruptionCount == 0 {
		t.Fatalf("expected non-zero corruption count; got 0")
	}
	if res.StaleReadCount != 0 {
		t.Fatalf("corruption-only mock should not trigger staleness; got stale=%d", res.StaleReadCount)
	}
}

func TestRunCacheProbe_DetectsStaleness(t *testing.T) {
	base := newBaseMock()
	base.seedTenant("t0", map[string][]byte{})
	stale := &staleMock{baseMock: base}
	res, err := RunCacheProbe(context.Background(), stale, "t0", cacheTestConcurrency, cacheTestDuration, cacheTestTTL, 50.0)
	t.Logf("seed=%d stale=%d corrupt=%d", res.Seed, res.StaleReadCount, res.CorruptionCount)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Pass {
		t.Fatalf("apparatus failed to detect staleness: %+v", res)
	}
	if res.StaleReadCount == 0 {
		t.Fatalf("expected non-zero stale read count; got 0")
	}
	if res.CorruptionCount != 0 {
		t.Fatalf("staleness-only mock should not trigger corruption; got corrupt=%d", res.CorruptionCount)
	}
	if len(res.StalePaths) == 0 {
		t.Fatalf("apparatus reported stale reads but did not record paths for diagnostic")
	}
}
