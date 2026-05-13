package mirage

// Coverage-drill-down tests authored to lift apparatus.go above the
// >90% line-coverage bar without resorting to a uniform "test
// everything" rule. Each function below corresponds to a bucket from
// docs/solutions/SOLUTION_COVERAGE_DRILL_DOWN.md:
//
//   Bucket 1 — Workspace-method error propagation: TestRun*_Propagates*
//   Bucket 2 — defensive input validation:         TestRun*_Rejects*, TestDecodeTimestamp_RejectsShortContent
//   Bucket 4 — missing-failure-mode diagnostic:    TestRunSnapshotProbe_DetectsMissingFilePostRestore
//
// Bucket 3 (freshSeed's crypto/rand fallback) is intentionally
// untested; see the doc comment on freshSeed for the structural
// reason.

import (
	"context"
	"strings"
	"testing"
	"time"
)

// --- Bucket 1: Workspace-method error propagation ---
//
// The apparatus's contract is that every probe wraps and returns
// underlying Workspace errors with a method-naming prefix. These
// tests assert the contract holds for each probe by feeding a mock
// that returns a sentinel error from the relevant method.

func TestRunOAuthProbe_PropagatesReadError(t *testing.T) {
	_, err := RunOAuthProbe(context.Background(), &erroringMock{}, []string{"t0"}, 1)
	if err == nil {
		t.Fatal("expected error propagation; got nil")
	}
	if !strings.Contains(err.Error(), "read") {
		t.Fatalf("expected wrapped read error; got: %v", err)
	}
}

func TestRunCacheProbe_PropagatesWriteError(t *testing.T) {
	_, err := RunCacheProbe(context.Background(), &erroringMock{}, "t0", 1, time.Millisecond, time.Second, 50.0)
	if err == nil {
		t.Fatal("expected error propagation; got nil")
	}
	if !strings.Contains(err.Error(), "pre-populate") {
		t.Fatalf("expected wrapped pre-populate error; got: %v", err)
	}
}

func TestRunListProbe_PropagatesListError(t *testing.T) {
	_, err := RunListProbe(context.Background(), &erroringMock{}, "t0", "", "", 0)
	if err == nil {
		t.Fatal("expected error propagation; got nil")
	}
	if !strings.Contains(err.Error(), "list") {
		t.Fatalf("expected wrapped list error; got: %v", err)
	}
}

func TestRunStatProbe_PropagatesStatError(t *testing.T) {
	_, err := RunStatProbe(context.Background(), &erroringMock{}, "t0", []string{"x"}, map[string]time.Time{})
	if err == nil {
		t.Fatal("expected error propagation; got nil")
	}
	if !strings.Contains(err.Error(), "stat") {
		t.Fatalf("expected wrapped stat error; got: %v", err)
	}
}

func TestRunSnapshotProbe_PropagatesPreManifestError(t *testing.T) {
	_, err := RunSnapshotProbe(context.Background(), &erroringMock{}, "t0", 0.1, time.Second, time.Second)
	if err == nil {
		t.Fatal("expected error propagation; got nil")
	}
	if !strings.Contains(err.Error(), "pre-snapshot manifest") {
		t.Fatalf("expected wrapped pre-snapshot manifest error; got: %v", err)
	}
}

func TestRunSnapshotProbe_PropagatesSnapshotError(t *testing.T) {
	base := newBaseMock()
	seedSnapshotTree(base, "t0")
	mock := &failingSnapshotMock{baseMock: base}
	_, err := RunSnapshotProbe(context.Background(), mock, "t0", 0.1, time.Second, time.Second)
	if err == nil {
		t.Fatal("expected error propagation; got nil")
	}
	if !strings.Contains(err.Error(), "snapshot") {
		t.Fatalf("expected wrapped snapshot error; got: %v", err)
	}
}

func TestRunSnapshotProbe_PropagatesRestoreError(t *testing.T) {
	base := newBaseMock()
	seedSnapshotTree(base, "t0")
	mock := &failingRestoreMock{baseMock: base}
	_, err := RunSnapshotProbe(context.Background(), mock, "t0", 0.1, time.Second, time.Second)
	if err == nil {
		t.Fatal("expected error propagation; got nil")
	}
	if !strings.Contains(err.Error(), "restore") {
		t.Fatalf("expected wrapped restore error; got: %v", err)
	}
}

func TestRunSnapshotProbe_PropagatesPostManifestError(t *testing.T) {
	base := newBaseMock()
	seedSnapshotTree(base, "t0")
	mock := &listCounterMock{baseMock: base}
	_, err := RunSnapshotProbe(context.Background(), mock, "t0", 0.1, time.Second, time.Second)
	if err == nil {
		t.Fatal("expected error propagation; got nil")
	}
	if !strings.Contains(err.Error(), "post-restore manifest") {
		t.Fatalf("expected wrapped post-restore manifest error; got: %v", err)
	}
}

// --- Bucket 2: defensive input validation ---

func TestRunOAuthProbe_RejectsEmptyTenants(t *testing.T) {
	_, err := RunOAuthProbe(context.Background(), newBaseMock(), nil, 10)
	if err == nil {
		t.Fatal("expected empty-tenants error; got nil")
	}
	if !strings.Contains(err.Error(), "tenants") {
		t.Fatalf("expected error naming the tenants guard; got: %v", err)
	}
}

func TestDecodeTimestamp_RejectsShortContent(t *testing.T) {
	if _, _, ok := decodeTimestamp([]byte{0x01, 0x02, 0x03}); ok {
		t.Fatal("expected ok=false for content shorter than timestamp prefix; got ok=true")
	}
}

// --- Bucket 4: missing-failure-mode diagnostic ---
//
// The "file missing post-restore" branch in RunSnapshotProbe is the
// failure shape a real backend would exhibit if Restore lost a file
// rather than corrupting one — a structurally distinct mode from
// byteDriftingMock. Without this test, the apparatus would be blind
// to one of the named diagnostic outputs it claims to produce.

func TestRunSnapshotProbe_DetectsMissingFilePostRestore(t *testing.T) {
	base := newBaseMock()
	seedSnapshotTree(base, "t0")
	mock := &restoreDropsFileMock{baseMock: base}
	res, err := RunSnapshotProbe(context.Background(), mock, "t0", 0.0, time.Second, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Pass {
		t.Fatalf("apparatus failed to detect dropped file: %+v", res)
	}
	if res.ByteIdentical {
		t.Fatal("expected ByteIdentical=false when a file is dropped post-restore")
	}
	var foundMissing bool
	for _, p := range res.DriftedPaths {
		if strings.Contains(p, "missing post-restore") {
			foundMissing = true
			break
		}
	}
	if !foundMissing {
		t.Fatalf("expected at least one path marked 'missing post-restore'; got %v", res.DriftedPaths)
	}
}
