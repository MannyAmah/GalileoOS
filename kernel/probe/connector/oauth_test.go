package connector

import (
	"context"
	"testing"
)

const oauthRequestCount = 1000 // synthetic-input scale; production gate uses 10000

func TestRunOAuthProbe_HappyPath(t *testing.T) {
	base := newBaseMock()
	for _, tenant := range []string{"a", "b", "c"} {
		base.seedTenant(tenant, map[string][]byte{"marker": []byte(tenant)})
	}
	res, err := RunOAuthProbe(context.Background(), base, []string{"a", "b", "c"}, oauthRequestCount)
	t.Logf("seed=%d (reproduce via go test -run TestRunOAuthProbe_HappyPath)", res.Seed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Pass {
		t.Fatalf("apparatus reported failure on happy-path mock: %+v", res)
	}
	if res.CorrectResponses != oauthRequestCount {
		t.Fatalf("expected %d correct responses, got %d", oauthRequestCount, res.CorrectResponses)
	}
	if res.CrossTenantLeaks != 0 {
		t.Fatalf("happy-path mock should produce 0 leaks, got %d", res.CrossTenantLeaks)
	}
}

func TestRunOAuthProbe_DetectsCrossTenantLeak(t *testing.T) {
	base := newBaseMock()
	for _, tenant := range []string{"a", "b", "c"} {
		base.seedTenant(tenant, map[string][]byte{"marker": []byte(tenant)})
	}
	leaking := &leakingMock{baseMock: base, leakRate: 0.10, rng: newFakeRand(0x1234567890ABCDEF)}
	res, err := RunOAuthProbe(context.Background(), leaking, []string{"a", "b", "c"}, oauthRequestCount)
	t.Logf("seed=%d", res.Seed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Pass {
		t.Fatalf("apparatus failed to detect cross-tenant leak: %+v", res)
	}
	if res.CrossTenantLeaks == 0 {
		t.Fatalf("expected at least one cross-tenant leak detected; got 0")
	}
	if len(res.LeakOccurrences) == 0 {
		t.Fatalf("apparatus reported leaks but did not record any occurrences for diagnostic output")
	}
}
