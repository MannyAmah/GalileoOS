package mirage

import (
	"context"
	"fmt"
	"testing"
	"time"
)

const (
	snapshotTestFileCount   = 32 // synthetic-input scale; production gate uses 100MB worth
	snapshotTestThreshold   = 25 * time.Millisecond
	snapshotTestMutateRatio = 0.10
)

func seedSnapshotTree(b *baseMock, tenant string) {
	files := map[string][]byte{}
	for i := 0; i < snapshotTestFileCount; i++ {
		files[fmt.Sprintf("file/%04d", i)] = []byte(fmt.Sprintf("content-%d-stable", i))
	}
	b.seedTenant(tenant, files)
}

func TestRunSnapshotProbe_HappyPath(t *testing.T) {
	base := newBaseMock()
	seedSnapshotTree(base, "t0")
	res, err := RunSnapshotProbe(context.Background(), base, "t0", snapshotTestMutateRatio, snapshotTestThreshold, snapshotTestThreshold)
	t.Logf("snap=%s restore=%s pre=%s post=%s", res.SnapshotDuration, res.RestoreDuration, res.PreHash[:12], res.PostHash[:12])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Pass {
		t.Fatalf("apparatus reported failure on happy-path mock: %+v", res)
	}
	if !res.ByteIdentical {
		t.Fatalf("expected byte-identical restore; got drift in %v", res.DriftedPaths)
	}
}

func TestRunSnapshotProbe_DetectsByteDrift(t *testing.T) {
	base := newBaseMock()
	seedSnapshotTree(base, "t0")
	drifting := &byteDriftingMock{baseMock: base}
	res, err := RunSnapshotProbe(context.Background(), drifting, "t0", snapshotTestMutateRatio, snapshotTestThreshold, snapshotTestThreshold)
	t.Logf("pre=%s post=%s drifted=%v", res.PreHash[:12], res.PostHash[:12], res.DriftedPaths)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Pass {
		t.Fatalf("apparatus failed to detect byte drift: %+v", res)
	}
	if res.ByteIdentical {
		t.Fatalf("expected ByteIdentical=false after byte flip; got true")
	}
	if len(res.DriftedPaths) == 0 {
		t.Fatalf("apparatus detected mismatch but did not record any drifted paths")
	}
}

func TestRunSnapshotProbe_DetectsSlowSnapshot(t *testing.T) {
	base := newBaseMock()
	seedSnapshotTree(base, "t0")
	slow := &slowSnapshotMock{baseMock: base}
	// Threshold 25ms; mock injects 50ms; must trip.
	res, err := RunSnapshotProbe(context.Background(), slow, "t0", snapshotTestMutateRatio, snapshotTestThreshold, snapshotTestThreshold)
	t.Logf("snap=%s threshold=%s", res.SnapshotDuration, snapshotTestThreshold)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Pass {
		t.Fatalf("apparatus failed to detect slow snapshot: %+v", res)
	}
	if res.SnapshotDuration < snapshotTestThreshold {
		t.Fatalf("expected SnapshotDuration >= threshold (%s); got %s", snapshotTestThreshold, res.SnapshotDuration)
	}
}
