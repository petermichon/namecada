// Package engine runs a set of providers concurrently and streams a result
// for each one as its check resolves, mirroring the frontend streamAll.
package engine

import (
	"context"
	"sync"
	"unicode/utf8"

	"github.com/petermichon/domain-browser/server/internal/provider"
	"github.com/petermichon/domain-browser/server/internal/store"
)

// ScanTimeout bounds a whole label's scan across every provider. It is twice
// the per-request timeout so a provider that makes two sequential calls (GitLab,
// Dev.to, OpenCollective) still fits, while a hung source cannot hold a scan
// open indefinitely.
const ScanTimeout = 2 * provider.DefaultTimeout

// DefaultConcurrency bounds how many provider checks run at once when an Engine
// is built. The registry has far more providers than this, so a scan of the
// default set proceeds in waves rather than opening one outbound check per
// provider at once.
const DefaultConcurrency = 8

// Result is the outcome of one provider's check for a label. Its shape
// mirrors the frontend ProviderResult so the two consumers stay in sync.
type Result struct {
	ID       string          `json:"id"`
	Platform string          `json:"platform"`
	Name     string          `json:"name"`
	URL      string          `json:"url"`
	Status   provider.Status `json:"status"`
	Detail   string          `json:"detail,omitempty"`
}

// Engine scans a label across a fixed, ordered set of providers.
type Engine struct {
	providers   []provider.Provider
	store       store.Store
	concurrency int
}

// New returns an Engine over the given providers, preserving their order. It
// starts with DefaultConcurrency checks in flight; use SetConcurrency to
// change that.
func New(providers ...provider.Provider) *Engine {
	return &Engine{
		providers:   append([]provider.Provider(nil), providers...),
		concurrency: DefaultConcurrency,
	}
}

// SetConcurrency caps how many provider checks the engine runs at once. A value
// of zero or less restores DefaultConcurrency. It must be called before Scan
// and is not safe to call concurrently with a scan.
func (e *Engine) SetConcurrency(n int) {
	if n <= 0 {
		n = DefaultConcurrency
	}
	e.concurrency = n
}

// NewWithStore returns an Engine over the given providers backed by s. Before
// checking, a provider's fresh cached entry is reused; a non-unknown result is
// then recorded, mirroring the frontend runProvider.
func NewWithStore(s store.Store, providers ...provider.Provider) *Engine {
	e := New(providers...)
	e.store = s
	return e
}

// NewFromRegistry returns an Engine over every provider in the registry.
func NewFromRegistry(r *provider.Registry) *Engine {
	return New(r.All()...)
}

// Providers returns the providers the engine will scan, in order.
func (e *Engine) Providers() []provider.Provider {
	return append([]provider.Provider(nil), e.providers...)
}

// Scan runs every provider through a bounded worker pool and returns a channel
// that emits each result as it resolves and then closes. At most the engine's
// concurrency (DefaultConcurrency unless SetConcurrency says otherwise) checks
// run at once, so a large registry does not open one outbound request per
// provider simultaneously; results still stream as each check finishes. A
// transport error is reported as an unknown status with an "unreachable"
// detail, matching the frontend. If ctx is canceled the engine stops emitting
// and closes the channel once the in-flight checks return.
func (e *Engine) Scan(ctx context.Context, label string) <-chan Result {
	out := make(chan Result, len(e.providers))

	workers := e.concurrency
	if workers < 1 {
		workers = DefaultConcurrency
	}
	if workers > len(e.providers) {
		workers = len(e.providers)
	}

	jobs := make(chan provider.Provider)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range jobs {
				res := e.check(ctx, p, label)
				if ctx.Err() != nil {
					continue
				}
				out <- res
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, p := range e.providers {
			select {
			case jobs <- p:
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// ScanAll collects Scan into a slice in provider order, mirroring the
// frontend checkAll.
func (e *Engine) ScanAll(ctx context.Context, label string) []Result {
	byID := make(map[string]Result)
	for res := range e.Scan(ctx, label) {
		byID[res.ID] = res
	}

	out := make([]Result, 0, len(e.providers))
	for _, p := range e.providers {
		if res, ok := byID[p.ID()]; ok {
			out = append(out, res)
		}
	}
	return out
}

func (e *Engine) check(ctx context.Context, p provider.Provider, label string) Result {
	res := Result{
		ID:       p.ID(),
		Platform: p.Platform(),
		Name:     p.Name(label),
		URL:      p.URL(label),
	}

	// A label shorter than the source's minimum can never exist there, so it
	// is reported taken without a request, mirroring the web app's runProvider.
	// Like there, the short-circuit result is not cached.
	if min := provider.MinLength(p); min > 0 && utf8.RuneCountInString(label) < min {
		res.Status = provider.StatusTaken
		return res
	}

	if e.store != nil {
		if entry, ok := e.store.Get(p.ID(), label); ok {
			res.Status = entry.Status
			res.Detail = entry.Detail
			return res
		}
	}

	got, err := p.Check(ctx, label)
	if err != nil {
		res.Status = provider.StatusUnknown
		res.Detail = "unreachable"
		return res
	}
	res.Status = got.Status
	res.Detail = got.Detail
	if e.store != nil && got.Status != provider.StatusUnknown {
		e.store.Set(p.ID(), label, got.Status, got.Detail)
	}
	return res
}
