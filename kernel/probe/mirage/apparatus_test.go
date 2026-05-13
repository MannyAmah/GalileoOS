package mirage

import (
	"context"
	"testing"
)

// TestComputeManifestHash_Deterministic verifies the manifest hash is
// stable across multiple calls on an unchanged tree, and changes when
// a single byte changes. Without this, the snapshot probe's drift
// detection can't be trusted.
func TestComputeManifestHash_Deterministic(t *testing.T) {
	base := newBaseMock()
	base.seedTenant("t0", map[string][]byte{
		"a": []byte("AAAA"),
		"b": []byte("BBBB"),
		"c": []byte("CCCC"),
	})
	h1, _, err := computeManifestHash(context.Background(), base, "t0")
	if err != nil {
		t.Fatalf("first hash: %v", err)
	}
	h2, _, err := computeManifestHash(context.Background(), base, "t0")
	if err != nil {
		t.Fatalf("second hash: %v", err)
	}
	if h1 != h2 {
		t.Fatalf("manifest hash not deterministic on unchanged tree: %s != %s", h1, h2)
	}
	// Flip one byte.
	if err := base.Write(context.Background(), "t0", "a", []byte("AAA!")); err != nil {
		t.Fatal(err)
	}
	h3, _, err := computeManifestHash(context.Background(), base, "t0")
	if err != nil {
		t.Fatalf("third hash: %v", err)
	}
	if h1 == h3 {
		t.Fatalf("manifest hash unchanged after byte flip; apparatus cannot detect drift")
	}
}

// TestFreshSeed_Unique sanity-checks that freshSeed returns distinct
// values across calls. Not a strict guarantee (the entropy source could
// in theory repeat), but a near-zero false positive rate.
func TestFreshSeed_Unique(t *testing.T) {
	seen := map[uint64]bool{}
	for i := 0; i < 16; i++ {
		s := freshSeed()
		if seen[s] {
			t.Fatalf("freshSeed returned duplicate %d on iteration %d", s, i)
		}
		seen[s] = true
	}
}
