package mirage

import (
	"context"
	"fmt"
	"testing"
)

func TestRunListProbe_HappyPath(t *testing.T) {
	base := newBaseMock()
	files := map[string][]byte{}
	for i := 0; i < 8; i++ {
		files[fmt.Sprintf("f-%d", i)] = []byte("x")
	}
	base.seedTenant("t0", files)
	res, err := RunListProbe(context.Background(), base, "t0", "", "" /*otherTenantPrefix*/, 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Pass {
		t.Fatalf("apparatus reported failure on happy-path mock: %+v", res)
	}
	if res.ActualCount != 8 {
		t.Fatalf("expected 8 entries; got %d", res.ActualCount)
	}
}

func TestRunListProbe_DetectsCrossTenantListLeak(t *testing.T) {
	base := newBaseMock()
	base.seedTenant("t0", map[string][]byte{"own": []byte("o")})
	base.seedTenant("other", map[string][]byte{"f-1": []byte("x"), "f-2": []byte("y")})
	leaking := &crossTenantListMock{baseMock: base, otherTenant: "other"}
	res, err := RunListProbe(context.Background(), leaking, "t0", "", "other", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Pass {
		t.Fatalf("apparatus failed to detect cross-tenant list leak: %+v", res)
	}
	if len(res.CrossTenantEntries) == 0 {
		t.Fatalf("apparatus reported failure but did not record cross-tenant entries for diagnostic")
	}
}
