package mirage

import (
	"context"
	"testing"
	"time"
)

func TestRunStatProbe_HappyPath(t *testing.T) {
	base := newBaseMock()
	base.seedTenant("t0", map[string][]byte{"a": []byte("1"), "b": []byte("22")})
	lastWrites := map[string]time.Time{"a": time.Now().Add(-time.Hour), "b": time.Now().Add(-time.Hour)}
	res, err := RunStatProbe(context.Background(), base, "t0", []string{"a", "b"}, lastWrites)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Pass {
		t.Fatalf("apparatus reported failure on happy-path mock: %+v", res)
	}
	if res.Calls != 2 {
		t.Fatalf("expected 2 calls; got %d", res.Calls)
	}
}

func TestRunStatProbe_DetectsStalenessMismatch(t *testing.T) {
	base := newBaseMock()
	base.seedTenant("t0", map[string][]byte{"a": []byte("1")})
	staleStat := &staleStatMock{baseMock: base}
	// Caller records the recent write; the mock returns ModTime=epoch
	// (definitely older). Apparatus must flag the mismatch.
	lastWrites := map[string]time.Time{"a": time.Now()}
	res, err := RunStatProbe(context.Background(), staleStat, "t0", []string{"a"}, lastWrites)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Pass {
		t.Fatalf("apparatus failed to detect stat staleness: %+v", res)
	}
	if len(res.StalenessMismatches) == 0 {
		t.Fatalf("apparatus reported failure but recorded no staleness mismatches")
	}
}
