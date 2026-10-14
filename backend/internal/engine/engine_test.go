package engine

import (
	"context"
	"testing"
	"time"

	"github.com/petermichon/domain-browser/backend/internal/provider"
)

// fake is a provider whose Check can block, fail, or report a fixed result.
type fake struct {
	id       string
	platform string
	result   provider.Result
	err      error
	release  <-chan struct{}
	started  chan<- string
}

func (f *fake) ID() string               { return f.id }
func (f *fake) Platform() string         { return f.platform }
func (f *fake) Name(label string) string { return f.id + ":" + label }
func (f *fake) URL(label string) string  { return "https://" + f.id + ".test/" + label }

func (f *fake) Check(ctx context.Context, label string) (provider.Result, error) {
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
