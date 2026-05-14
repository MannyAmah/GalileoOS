package connector

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// baseMock is the happy-path mock Workspace used as the apparatus's
// known-good input. Failure-injection mocks embed it and override the
// one method they want to corrupt.
type baseMock struct {
	mu            sync.Mutex
	tree          map[string]map[string][]byte // tenant -> path -> content
	writes        map[string]map[string]time.Time
	snapshots     map[string]map[string]map[string][]byte // token -> tenant -> path -> content
	readLatencyNs int64
}

func newBaseMock() *baseMock {
	return &baseMock{
		tree:      map[string]map[string][]byte{},
		writes:    map[string]map[string]time.Time{},
		snapshots: map[string]map[string]map[string][]byte{},
	}
}

func (m *baseMock) seedTenant(tenant string, files map[string][]byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.tree[tenant] == nil {
		m.tree[tenant] = map[string][]byte{}
		m.writes[tenant] = map[string]time.Time{}
	}
	for k, v := range files {
		m.tree[tenant][k] = append([]byte(nil), v...)
		m.writes[tenant][k] = time.Now()
	}
}

func (m *baseMock) List(_ context.Context, tenant, _ string) ([]Entry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Entry, 0, len(m.tree[tenant]))
	for name, content := range m.tree[tenant] {
		out = append(out, Entry{Name: name, Size: int64(len(content)), ModTime: m.writes[tenant][name], ContentType: "application/octet-stream"})
	}
	return out, nil
}

func (m *baseMock) Stat(_ context.Context, tenant, path string) (Metadata, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	content, ok := m.tree[tenant][path]
	if !ok {
		return Metadata{}, fmt.Errorf("not found: %s/%s", tenant, path)
	}
	return Metadata{Name: path, Size: int64(len(content)), ModTime: m.writes[tenant][path], ContentType: "application/octet-stream", ETag: fmt.Sprintf("%x", len(content))}, nil
}

func (m *baseMock) Read(_ context.Context, tenant, path string) ([]byte, error) {
	if m.readLatencyNs > 0 {
		time.Sleep(time.Duration(m.readLatencyNs))
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	content, ok := m.tree[tenant][path]
	if !ok {
		return nil, fmt.Errorf("not found: %s/%s", tenant, path)
	}
	return append([]byte(nil), content...), nil
}

func (m *baseMock) Write(_ context.Context, tenant, path string, content []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.tree[tenant] == nil {
		m.tree[tenant] = map[string][]byte{}
		m.writes[tenant] = map[string]time.Time{}
	}
	m.tree[tenant][path] = append([]byte(nil), content...)
	m.writes[tenant][path] = time.Now()
	return nil
}

func (m *baseMock) Snapshot(_ context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	token := fmt.Sprintf("snap-%d", time.Now().UnixNano())
	snap := map[string]map[string][]byte{}
	for tenant, files := range m.tree {
		snap[tenant] = map[string][]byte{}
		for p, c := range files {
			snap[tenant][p] = append([]byte(nil), c...)
		}
	}
	m.snapshots[token] = snap
	return token, nil
}

func (m *baseMock) Restore(_ context.Context, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	snap, ok := m.snapshots[token]
	if !ok {
		return fmt.Errorf("unknown snapshot token: %s", token)
	}
	m.tree = map[string]map[string][]byte{}
	m.writes = map[string]map[string]time.Time{}
	for tenant, files := range snap {
		m.tree[tenant] = map[string][]byte{}
		m.writes[tenant] = map[string]time.Time{}
		for p, c := range files {
			m.tree[tenant][p] = append([]byte(nil), c...)
			m.writes[tenant][p] = time.Now()
		}
	}
	return nil
}

// --- failure-injection mocks ---

// leakingMock returns tenant A's marker when tenant B asks, at the
// configured leakRate (per request). The apparatus must report
// "cross-tenant leak: N occurrences."
type leakingMock struct {
	*baseMock
	leakRate float32 // 0.0 .. 1.0
	rng      *fakeRand
}

func (m *leakingMock) Read(ctx context.Context, tenant, path string) ([]byte, error) {
	if m.rng.next() < m.leakRate {
		// Return some *other* tenant's content.
		m.mu.Lock()
		for other, files := range m.tree {
			if other != tenant {
				if c, ok := files[path]; ok {
					m.mu.Unlock()
					return append([]byte(nil), c...), nil
				}
			}
		}
		m.mu.Unlock()
	}
	return m.baseMock.Read(ctx, tenant, path)
}

// slowMock injects a 10ms read delay so tests with a 5ms p99 threshold
// can trip the apparatus's pass-criterion in under a second of
// wall-clock instead of the >>100ms needed against the production
// 100ms threshold. Production gate criterion is still 100ms; the
// threshold is supplied by the caller.
type slowMock struct{ *baseMock }

func (m *slowMock) Read(ctx context.Context, tenant, path string) ([]byte, error) {
	time.Sleep(10 * time.Millisecond)
	return m.baseMock.Read(ctx, tenant, path)
}

// corruptingMock flips one byte in the SUFFIX (content beyond the
// timestamp prefix) on every read. The apparatus's SHA-256 check
// computes over the suffix, so this mock triggers corruption WITHOUT
// triggering staleness.
type corruptingMock struct{ *baseMock }

func (m *corruptingMock) Read(ctx context.Context, tenant, path string) ([]byte, error) {
	b, err := m.baseMock.Read(ctx, tenant, path)
	if err == nil && len(b) > 8 {
		b[8] ^= 0xFF // flip a byte in the suffix only
	}
	return b, err
}

// staleMock rewrites the timestamp PREFIX on read to epoch (0). The
// apparatus decodes the prefix to check freshness; the suffix is
// untouched so SHA-256 corruption check passes. This mock therefore
// triggers staleness WITHOUT triggering corruption — orthogonal
// failure modes.
type staleMock struct{ *baseMock }

func (m *staleMock) Read(ctx context.Context, tenant, path string) ([]byte, error) {
	b, err := m.baseMock.Read(ctx, tenant, path)
	if err == nil && len(b) >= 8 {
		// Zero the 8-byte timestamp prefix -> looks ~58 years stale.
		for i := 0; i < 8; i++ {
			b[i] = 0
		}
	}
	return b, err
}

// byteDriftingMock's Restore flips one byte in one file post-restore.
type byteDriftingMock struct{ *baseMock }

func (m *byteDriftingMock) Restore(ctx context.Context, token string) error {
	if err := m.baseMock.Restore(ctx, token); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, files := range m.tree {
		for p := range files {
			if len(files[p]) > 0 {
				files[p][0] ^= 0xFF
				return nil
			}
			_ = p
		}
	}
	return nil
}

// slowSnapshotMock injects 50ms delay into Snapshot so tests can use a
// short threshold (e.g., 25ms) and trip the apparatus's pass-criterion
// in under a second of wall-clock instead of 10s+. Production gate
// criterion is still 10s; the threshold is supplied by the caller.
type slowSnapshotMock struct{ *baseMock }

func (m *slowSnapshotMock) Snapshot(ctx context.Context) (string, error) {
	time.Sleep(50 * time.Millisecond)
	return m.baseMock.Snapshot(ctx)
}

// crossTenantListMock returns tenant B's entries when tenant A lists.
type crossTenantListMock struct {
	*baseMock
	otherTenant string
}

func (m *crossTenantListMock) List(ctx context.Context, tenant, prefix string) ([]Entry, error) {
	// Returns the wrong tenant's tree but tags entries so the probe
	// can detect the leak by name prefix.
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Entry, 0, len(m.tree[m.otherTenant]))
	for name := range m.tree[m.otherTenant] {
		out = append(out, Entry{Name: m.otherTenant + "/" + name, Size: 1, ModTime: time.Now(), ContentType: "application/octet-stream"})
	}
	return out, nil
}

// staleStatMock returns Stat ModTime = epoch (definitely older than any
// recent Write recorded by the caller).
type staleStatMock struct{ *baseMock }

func (m *staleStatMock) Stat(ctx context.Context, tenant, path string) (Metadata, error) {
	md, err := m.baseMock.Stat(ctx, tenant, path)
	if err != nil {
		return md, err
	}
	md.ModTime = time.Unix(0, 0)
	return md, nil
}

// erroringMock returns errSentinel from every Workspace method. Used
// by Bucket-1 (error-propagation) coverage tests to assert that every
// probe wraps and returns the underlying Workspace error with a
// method-naming prefix rather than swallowing it.
type erroringMock struct{}

var errSentinel = fmt.Errorf("simulated workspace failure")

func (m *erroringMock) List(_ context.Context, _, _ string) ([]Entry, error) {
	return nil, errSentinel
}
func (m *erroringMock) Stat(_ context.Context, _, _ string) (Metadata, error) {
	return Metadata{}, errSentinel
}
func (m *erroringMock) Read(_ context.Context, _, _ string) ([]byte, error) {
	return nil, errSentinel
}
func (m *erroringMock) Write(_ context.Context, _, _ string, _ []byte) error {
	return errSentinel
}
func (m *erroringMock) Snapshot(_ context.Context) (string, error) {
	return "", errSentinel
}
func (m *erroringMock) Restore(_ context.Context, _ string) error {
	return errSentinel
}

// failingSnapshotMock works correctly except Snapshot returns error.
// Exercises RunSnapshotProbe's Snapshot-failure branch in isolation.
type failingSnapshotMock struct{ *baseMock }

func (m *failingSnapshotMock) Snapshot(_ context.Context) (string, error) {
	return "", fmt.Errorf("simulated snapshot failure")
}

// failingRestoreMock works correctly except Restore returns error.
// Exercises RunSnapshotProbe's Restore-failure branch in isolation.
type failingRestoreMock struct{ *baseMock }

func (m *failingRestoreMock) Restore(_ context.Context, _ string) error {
	return fmt.Errorf("simulated restore failure")
}

// listCounterMock's List succeeds on the first call (pre-snapshot
// walk) and fails on the second (post-restore walk). Exercises
// RunSnapshotProbe's post-restore-manifest error branch without
// requiring a second mock type for nearly-identical behavior.
type listCounterMock struct {
	*baseMock
	calls int
}

func (m *listCounterMock) List(ctx context.Context, tenant, prefix string) ([]Entry, error) {
	m.mu.Lock()
	m.calls++
	c := m.calls
	m.mu.Unlock()
	if c >= 2 {
		return nil, fmt.Errorf("simulated post-restore list failure")
	}
	return m.baseMock.List(ctx, tenant, prefix)
}

// restoreDropsFileMock's Restore omits one file after a successful
// underlying restore — i.e., the restored set is a strict subset of
// the snapshot set. Triggers RunSnapshotProbe's "missing post-restore"
// diagnostic branch, which is the failure shape a real backend would
// exhibit if Restore lost a file rather than corrupting one.
type restoreDropsFileMock struct{ *baseMock }

func (m *restoreDropsFileMock) Restore(ctx context.Context, token string) error {
	if err := m.baseMock.Restore(ctx, token); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, files := range m.tree {
		for p := range files {
			delete(files, p)
			return nil
		}
	}
	return nil
}

// fakeRand is a tiny deterministic float source for failure-injection
// mocks that need reproducible bad behavior under a known seed.
type fakeRand struct {
	state uint64
}

func newFakeRand(seed uint64) *fakeRand { return &fakeRand{state: seed} }

func (r *fakeRand) next() float32 {
	// xorshift64* — adequate for "is this request in the leak window".
	r.state ^= r.state << 13
	r.state ^= r.state >> 7
	r.state ^= r.state << 17
	return float32(r.state%1000) / 1000.0
}
