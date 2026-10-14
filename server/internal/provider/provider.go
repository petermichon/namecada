// Package provider defines the availability-provider contract and the
// registry that the scan engine iterates.
package provider

import (
	"context"
	"errors"
	"fmt"
	"net/url"
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

// base carries the identity and presentation shared by every provider that
// needs a hand-written Check. Embedding it promotes ID, Platform, Name and URL,
// so such a provider only has to implement Check.
type base struct {
	id       string
	platform string
	// nameFmt formats the human label; the empty string keeps the label as-is.
	nameFmt string
	// urlFmt formats the public URL, with one %s for the label.
	urlFmt string
	// rawURL inserts the label into urlFmt unescaped; the default escapes it.
	rawURL bool
}

func (b base) ID() string       { return b.id }
func (b base) Platform() string { return b.platform }

func (b base) Name(label string) string {
	if b.nameFmt == "" {
		return label
	}
	return fmt.Sprintf(b.nameFmt, label)
}

func (b base) URL(label string) string {
	if b.rawURL {
		return fmt.Sprintf(b.urlFmt, label)
	}
	return fmt.Sprintf(b.urlFmt, url.PathEscape(label))
}

// MinLengthProvider is implemented by a provider that rejects labels shorter
// than a minimum length. It is optional: a provider that does not implement it
// still inherits any minimum recorded for its id in the known table.
type MinLengthProvider interface {
	// MinLength is the shortest label the source accepts; 0 means no minimum.
	MinLength() int
}

// minLengths records the known minimum label length for sources whose rules
// are documented. A label shorter than its provider's minimum cannot exist on
// that source, so the engine reports it as taken without a request.
var minLengths = map[string]int{
	"github":      2,
	"lemmy":       2,
	"hex":         2,
	"hex-user":    3,
	"wordpress":   5,
	"bitbucket":   4,
	"minehut":     4,
	"lichess":     2,
	"chesscom":    3,
	"huggingface": 2,
	"dailymotion": 3,
	"bluesky":     3,
	"devto":       2,
}

// MinLength reports the minimum label length for p. An explicit
// MinLengthProvider wins; otherwise the known table is consulted by id. A
// result of 0 means the provider has no minimum.
func MinLength(p Provider) int {
	if p == nil {
		return 0
	}
	if ml, ok := p.(MinLengthProvider); ok {
		return ml.MinLength()
	}
	return minLengths[p.ID()]
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
	for _, p := range NewDomains() {
		_ = r.Register(p)
	}
	_ = r.Register(NewLemmy())
	_ = r.Register(NewPeerTube())
	_ = r.Register(NewNPM())
	_ = r.Register(NewPyPI())
	_ = r.Register(NewCrates())
	_ = r.Register(NewNuGet())
	_ = r.Register(NewHex())
	_ = r.Register(NewHexUser())
	_ = r.Register(NewPub())
	_ = r.Register(NewHomebrew())
	_ = r.Register(NewFirefox())
	_ = r.Register(NewJSR())
	_ = r.Register(NewWordPress())
	_ = r.Register(NewOpenVSX())
	_ = r.Register(NewGitHub())
	_ = r.Register(NewGitLab())
	_ = r.Register(NewBitbucket())
	_ = r.Register(NewCodeberg())
	_ = r.Register(NewWikipedia())
	_ = r.Register(NewDevTo())
	_ = r.Register(NewFarcaster())
	_ = r.Register(NewMinehut())
	_ = r.Register(NewLichess())
	_ = r.Register(NewChesscom())
	_ = r.Register(NewHuggingFace())
	_ = r.Register(NewDailymotion())
	_ = r.Register(NewMixcloud())
	_ = r.Register(NewBluesky())
	_ = r.Register(NewMastodon())
	_ = r.Register(NewOpenCollective())
	return r
}
