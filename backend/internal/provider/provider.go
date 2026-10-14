// Package provider defines the availability-provider contract and the
// registry that the scan engine iterates.
package provider

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Status is the outcome of a single availability check.
type Status string

const (
	StatusAvailable Status = "available"
	StatusTaken     Status = "taken"
	StatusUnknown   Status = "unknown"
)

// Result is what a provider reports for a label.
type Result struct {
	Status Status
	Detail string
}

// Provider checks whether a label is taken on one source.
type Provider interface {
	// ID is the stable, unique identifier of the provider (e.g. "lemmy").
	ID() string
	// Platform groups providers that share a namespace (e.g. "lemmy").
	Platform() string
	// Name is the human label a result is shown under.
	Name(label string) string
	// URL is the public page for the label.
	URL(label string) string
	// Check reports whether the label is taken, available, or unknown.
	Check(ctx context.Context, label string) (Result, error)
}

// ErrDuplicate is returned when a provider id is registered twice.
var ErrDuplicate = errors.New("provider: duplicate id")

// Registry is an ordered, id-addressable set of providers.
type Registry struct {
	mu    sync.RWMutex
	order []string
	byID  map[string]Provider
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{byID: make(map[string]Provider)}
}

// Register adds a provider. It rejects nil providers and duplicate ids.
func (r *Registry) Register(p Provider) error {
	if p == nil {
		return errors.New("provider: cannot register nil")
	}
	id := p.ID()
	if id == "" {
		return errors.New("provider: id must not be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byID[id]; ok {
		return fmt.Errorf("%w: %s", ErrDuplicate, id)
	}
	r.byID[id] = p
	r.order = append(r.order, id)
	return nil
}

// Get returns the provider with the given id.
func (r *Registry) Get(id string) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.byID[id]
	return p, ok
}

// All returns the providers in registration order.
func (r *Registry) All() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Provider, 0, len(r.order))
	for _, id := range r.order {
		out = append(out, r.byID[id])
	}
	return out
}

// Len reports how many providers are registered.
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.order)
}

// Default returns a registry populated with the built-in providers.
func Default() *Registry {
	r := NewRegistry()
	// Registration cannot fail for the constants below.
	_ = r.Register(NewLemmy())
	_ = r.Register(NewPeerTube())
	return r
}
