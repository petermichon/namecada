package engine

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/petermichon/domain-browser/server/internal/provider"
	"github.com/petermichon/domain-browser/server/internal/store"
)

// fake is a provider whose Check can block, fail, or report a fixed result.
type fake struct {
	id        string
	platform  string
	minLength int
	result    provider.Result
	err       error
	release   <-chan struct{}
	started   chan<- string
	calls     *atomic.Int32
	active    *atomic.Int32
	max       *atomic.Int32
}

func (f *fake) ID() string               { return f.id }
func (f *fake) Platform() string         { return f.platform }
func (f *fake) Name(label string) string { return f.id + ":" + label }
func (f *fake) URL(label string) string  { return "https://" + f.id + ".test/" + label }
func (f *fake) MinLength() int           { return f.minLength }

func (f *fake) Check(ctx context.Context, label string) (provider.Result, error) {
	if f.calls != nil {
		f.calls.Add(1)
	}
	if f.active != nil {
		cur := f.active.Add(1)
		defer f.active.Add(-1)
		if f.max != nil {
			for {
				peak := f.max.Load()
				if cur <= peak || f.max.CompareAndSwap(peak, cur) {
					break
				}
			}
		}
	}
	if f.started != nil {
		f.started <- f.id
	}
	if f.release != nil {
		select {
		case <-f.release:
		case <-ctx.Done():
			return provider.Result{}, ctx.Err()
		}
	}
	return f.result, f.err
}

func waitClosed(t *testing.T, ch <-chan Result) []Result {
	t.Helper()
	var got []Result
	timeout := time.After(2 * time.Second)
	for {
		select {
		case res, ok := <-ch:
			if !ok {
				return got
			}
			got = append(got, res)
		case <-timeout:
			t.Fatal("Scan did not close within the timeout")
		}
	}
}

func TestScanEmitsOneResultPerProvider(t *testing.T) {
	a := &fake{id: "a", platform: "pa", result: provider.Result{Status: provider.StatusAvailable, Detail: "free"}}
	b := &fake{id: "b", platform: "pb", result: provider.Result{Status: provider.StatusTaken}}

	got := waitClosed(t, New(a, b).Scan(context.Background(), "acema"))
	if len(got) != 2 {
		t.Fatalf("got %d results, want 2: %+v", len(got), got)
	}

	byID := map[string]Result{}
	for _, r := range got {
		byID[r.ID] = r
	}

	if r := byID["a"]; r.Status != provider.StatusAvailable || r.Detail != "free" {
		t.Errorf("a = %+v", r)
	}
	if r := byID["b"]; r.Status != provider.StatusTaken {
		t.Errorf("b = %+v", r)
	}

	wantURL := "https://a.test/acema"
	if r := byID["a"]; r.Platform != "pa" || r.Name != "a:acema" || r.URL != wantURL {
		t.Errorf("a metadata = %+v, want platform pa name a:acema url %s", r, wantURL)
	}
}

func TestScanPopulatesPlatformName(t *testing.T) {
	p := &fake{id: "pypi", platform: "pypi", result: provider.Result{Status: provider.StatusAvailable}}
	unknown := &fake{id: "other", platform: "brand-new", result: provider.Result{Status: provider.StatusTaken}}

	byID := map[string]Result{}
	for _, r := range New(p, unknown).ScanAll(context.Background(), "acema") {
		byID[r.ID] = r
	}
	if got := byID["pypi"].PlatformName; got != "PyPI" {
		t.Errorf("pypi platformName = %q, want %q", got, "PyPI")
	}
	if got := byID["other"].PlatformName; got != "brand-new" {
		t.Errorf("unknown platformName = %q, want the id unchanged", got)
	}
}

func TestScanRunsProvidersConcurrently(t *testing.T) {
	const n = 5
	started := make(chan string, n)
	release := make(chan struct{})
	providers := make([]provider.Provider, n)
	for i := 0; i < n; i++ {
		providers[i] = &fake{
			id:      string(rune('a' + i)),
			release: release,
			started: started,
		}
	}

	ch := New(providers...).Scan(context.Background(), "acema")

	for i := 0; i < n; i++ {
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			t.Fatalf("only %d of %d providers started; engine is not concurrent", i, n)
		}
	}
	close(release)

	if got := waitClosed(t, ch); len(got) != n {
		t.Fatalf("got %d results, want %d", len(got), n)
	}
}

func TestScanBoundsConcurrentChecks(t *testing.T) {
	const n, limit = 6, 2
	release := make(chan struct{})
	started := make(chan string, n)
	var active, max atomic.Int32
	providers := make([]provider.Provider, n)
	for i := 0; i < n; i++ {
		providers[i] = &fake{
			id:      string(rune('a' + i)),
			release: release,
			started: started,
			active:  &active,
			max:     &max,
		}
	}

	eng := New(providers...)
	eng.SetConcurrency(limit)
	ch := eng.Scan(context.Background(), "acema")

	for i := 0; i < limit; i++ {
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			t.Fatalf("only %d of %d providers started", i, limit)
		}
	}
	select {
	case id := <-started:
		t.Fatalf("provider %q started while %d checks were in flight; cap is %d", id, limit, limit)
	case <-time.After(100 * time.Millisecond):
	}

	close(release)
	if got := waitClosed(t, ch); len(got) != n {
		t.Fatalf("got %d results, want %d", len(got), n)
	}
	if got := max.Load(); got > limit {
		t.Fatalf("peak concurrency = %d, want <= %d", got, limit)
	}
}

func TestScanSetConcurrencyZeroUsesDefault(t *testing.T) {
	const n = 5
	started := make(chan string, n)
	release := make(chan struct{})
	providers := make([]provider.Provider, n)
	for i := 0; i < n; i++ {
		providers[i] = &fake{
			id:      string(rune('a' + i)),
			release: release,
			started: started,
		}
	}

	eng := New(providers...)
	eng.SetConcurrency(0)
	ch := eng.Scan(context.Background(), "acema")

	if DefaultConcurrency < n {
		t.Fatalf("test assumes DefaultConcurrency (%d) >= %d", DefaultConcurrency, n)
	}
	for i := 0; i < n; i++ {
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			t.Fatalf("only %d of %d providers started; a zero cap should restore the default", i, n)
		}
	}
	close(release)

	if got := waitClosed(t, ch); len(got) != n {
		t.Fatalf("got %d results, want %d", len(got), n)
	}
}

func TestScanStreamsAsEachResolves(t *testing.T) {
	release := make(chan struct{})
	slow := &fake{id: "slow", release: release, result: provider.Result{Status: provider.StatusTaken}}
	fast := &fake{id: "fast", result: provider.Result{Status: provider.StatusAvailable}}

	ch := New(slow, fast).Scan(context.Background(), "acema")

	select {
	case first := <-ch:
		if first.ID != "fast" {
			t.Fatalf("first result = %q, want the fast provider", first.ID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no result streamed before the slow provider finished")
	}

	close(release)
	if got := waitClosed(t, ch); len(got) != 1 || got[0].ID != "slow" {
		t.Fatalf("remaining results = %+v, want [slow]", got)
	}
}

func TestScanMapsErrorsToUnknown(t *testing.T) {
	boom := &fake{id: "boom", err: context.DeadlineExceeded}
	ok := &fake{id: "ok", result: provider.Result{Status: provider.StatusAvailable}}

	got := waitClosed(t, New(boom, ok).Scan(context.Background(), "acema"))
	if len(got) != 2 {
		t.Fatalf("got %d results, want 2", len(got))
	}
	for _, r := range got {
		if r.ID == "boom" {
			if r.Status != provider.StatusUnknown || r.Detail != "unreachable" {
				t.Errorf("boom = %+v, want unknown/unreachable", r)
			}
		}
	}
}

func TestScanTreatsShortLabelsAsTaken(t *testing.T) {
	var calls atomic.Int32
	p := &fake{
		id:        "a",
		minLength: 4,
		result:    provider.Result{Status: provider.StatusAvailable, Detail: "free"},
		calls:     &calls,
	}

	short := New(p).ScanAll(context.Background(), "abc")
	if len(short) != 1 || short[0].Status != provider.StatusTaken {
		t.Fatalf("short label = %+v, want a single taken result", short)
	}
	if calls.Load() != 0 {
		t.Fatalf("Check calls = %d, want 0 (a short label short-circuits)", calls.Load())
	}

	atMinimum := New(p).ScanAll(context.Background(), "abcd")
	if len(atMinimum) != 1 || atMinimum[0].Status != provider.StatusAvailable {
		t.Fatalf("at-minimum label = %+v, want available", atMinimum)
	}
	if calls.Load() != 1 {
		t.Fatalf("Check calls = %d, want 1 after the at-minimum scan", calls.Load())
	}
}

func TestScanDoesNotCacheShortLabels(t *testing.T) {
	var calls atomic.Int32
	p := &fake{id: "a", minLength: 4, result: provider.Result{Status: provider.StatusAvailable}, calls: &calls}
	mem := store.NewMemory()
	eng := NewWithStore(mem, p)

	if got := eng.ScanAll(context.Background(), "abc"); len(got) != 1 || got[0].Status != provider.StatusTaken {
		t.Fatalf("short label = %+v, want taken", got)
	}
	if _, ok := mem.Get("a", "abc"); ok {
		t.Fatal("short label was cached, want not cached")
	}

	if got := eng.ScanAll(context.Background(), "abcd"); len(got) != 1 || got[0].Status != provider.StatusAvailable {
		t.Fatalf("at-minimum label = %+v, want available", got)
	}
	if calls.Load() != 1 {
		t.Fatalf("Check calls = %d, want 1", calls.Load())
	}
}

func TestScanAllPreservesProviderOrder(t *testing.T) {
	release := make(chan struct{})
	a := &fake{id: "a", release: release, result: provider.Result{Status: provider.StatusTaken}}
	b := &fake{id: "b", release: release, result: provider.Result{Status: provider.StatusAvailable}}
	c := &fake{id: "c", result: provider.Result{Status: provider.StatusTaken}}
	close(release)

	got := New(a, b, c).ScanAll(context.Background(), "acema")
	if len(got) != 3 {
		t.Fatalf("got %d results, want 3", len(got))
	}
	for i, want := range []string{"a", "b", "c"} {
		if got[i].ID != want {
			t.Fatalf("result[%d] = %q, want %q (order %v)", i, got[i].ID, want, ids(got))
		}
	}
}

func TestScanAllReusesStoreCache(t *testing.T) {
	var calls atomic.Int32
	p := &fake{
		id:     "a",
		result: provider.Result{Status: provider.StatusTaken, Detail: "found"},
		calls:  &calls,
	}
	eng := NewWithStore(store.NewMemory(), p)

	first := eng.ScanAll(context.Background(), "acema")
	if len(first) != 1 || first[0].Status != provider.StatusTaken || first[0].Detail != "found" {
		t.Fatalf("first = %+v, want taken/found", first)
	}
	if calls.Load() != 1 {
		t.Fatalf("calls after first scan = %d, want 1", calls.Load())
	}

	second := eng.ScanAll(context.Background(), "acema")
	if len(second) != 1 || second[0].Status != provider.StatusTaken || second[0].Detail != "found" {
		t.Fatalf("second = %+v, want the cached taken/found", second)
	}
	if calls.Load() != 1 {
		t.Fatalf("calls after cached scan = %d, want 1", calls.Load())
	}
}

func TestScanAllDoesNotCacheUnknown(t *testing.T) {
	var calls atomic.Int32
	p := &fake{id: "a", result: provider.Result{Status: provider.StatusUnknown}, calls: &calls}
	eng := NewWithStore(store.NewMemory(), p)

	eng.ScanAll(context.Background(), "acema")
	eng.ScanAll(context.Background(), "acema")
	if calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2 (unknown results are not cached)", calls.Load())
	}
}

func TestScanAllDoesNotCacheErrors(t *testing.T) {
	var calls atomic.Int32
	p := &fake{id: "a", err: context.DeadlineExceeded, calls: &calls}
	eng := NewWithStore(store.NewMemory(), p)

	got := eng.ScanAll(context.Background(), "acema")
	if len(got) != 1 || got[0].Status != provider.StatusUnknown || got[0].Detail != "unreachable" {
		t.Fatalf("got %+v, want unknown/unreachable", got)
	}
	if calls.Load() != 1 {
		t.Fatalf("calls after first scan = %d, want 1", calls.Load())
	}

	eng.ScanAll(context.Background(), "acema")
	if calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2 (errors are not cached)", calls.Load())
	}
}

func TestScanStopsEmittingAfterCancel(t *testing.T) {
	started := make(chan string, 1)
	release := make(chan struct{})
	block := &fake{id: "block", started: started, release: release}
	eng := New(block)

	ctx, cancel := context.WithCancel(context.Background())
	ch := eng.Scan(ctx, "acema")

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("provider never started")
	}
	cancel()

	if got := waitClosed(t, ch); len(got) != 0 {
		t.Fatalf("got %+v after cancel, want no results", got)
	}
}

func TestNewFromRegistry(t *testing.T) {
	r := provider.NewRegistry()
	if err := r.Register(&fake{id: "a"}); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(&fake{id: "b"}); err != nil {
		t.Fatal(err)
	}

	got := NewFromRegistry(r).Providers()
	if idsFrom(got) != "a,b" {
		t.Fatalf("providers = %v, want [a b]", idsFrom(got))
	}
}

func ids(rs []Result) []string {
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = r.ID
	}
	return out
}

func idsFrom(ps []provider.Provider) string {
	out := ""
	for i, p := range ps {
		if i > 0 {
			out += ","
		}
		out += p.ID()
	}
	return out
}
