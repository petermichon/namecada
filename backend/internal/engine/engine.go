// Package engine runs a set of providers concurrently and streams a result
// for each one as its check resolves, mirroring the frontend streamAll.
package engine

import (
	"context"
	"sync"

	"github.com/petermichon/domain-browser/backend/internal/provider"
	"github.com/petermichon/domain-browser/backend/internal/store"
)

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
	providers []provider.Provider
	store     store.Store
}

// New returns an Engine over the given providers, preserving their order.
func New(providers ...provider.Provider) *Engine {
	return &Engine{providers: append([]provider.Provider(nil), providers...)}
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

// Scan runs every provider concurrently and returns a channel that emits each
// result as it resolves and then closes. A transport error is reported as an
// unknown status with an "unreachable" detail, matching the frontend. If ctx
// is canceled the engine stops emitting and closes the channel once the
// in-flight checks return.
func (e *Engine) Scan(ctx context.Context, label string) <-chan Result {
	out := make(chan Result, len(e.providers))

	var wg sync.WaitGroup
	for _, p := range e.providers {
		wg.Add(1)
		go func(p provider.Provider) {
			defer wg.Done()
			res := e.check(ctx, p, label)
			if ctx.Err() != nil {
				return
			}
			out <- res
		}(p)
	}

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
