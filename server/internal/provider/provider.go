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
	// URL is the matched public page when it differs from the provider's
	// default URL(label); empty means the default applies.
	URL string
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
	// paid marks a source that charges money to register a name (domains).
	paid bool
}

func (b base) ID() string       { return b.id }
func (b base) Platform() string { return b.platform }
func (b base) Paid() bool       { return b.paid }

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
	"lemmy":       3,
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

// MaxLengthProvider is implemented by a provider that rejects labels longer
// than a maximum length. It is optional: a provider that does not implement it
// still inherits any maximum recorded for its id in the known table.
type MaxLengthProvider interface {
	// MaxLength is the longest label the source accepts; 0 means no maximum.
	MaxLength() int
}

// maxLengths records the known maximum label length for sources whose rules
// are documented. A label longer than its provider's maximum cannot exist on
// that source, so the engine reports it as taken without a request.
var maxLengths = map[string]int{
	"lemmy": 20,
}

// MaxLength reports the maximum label length for p. An explicit
// MaxLengthProvider wins; otherwise the known table is consulted by id. A
// result of 0 means the provider has no maximum.
func MaxLength(p Provider) int {
	if p == nil {
		return 0
	}
	if ml, ok := p.(MaxLengthProvider); ok {
		return ml.MaxLength()
	}
	return maxLengths[p.ID()]
}

// PaidProvider is implemented by a provider whose names cost money to
// register. It is optional: a provider that does not implement it, or that
// reports false, is free.
type PaidProvider interface {
	// Paid reports whether registering a name on the source costs money.
	Paid() bool
}

// Paid reports whether p costs money to register. A nil or non-implementing
// provider is free.
func Paid(p Provider) bool {
	if p == nil {
		return false
	}
	if pp, ok := p.(PaidProvider); ok {
		return pp.Paid()
	}
	return false
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

// Config carries the credentials a credential-gated ("keyed") source needs. It
// is populated once at process start (see cmd/domainbrowser) from
// DOMAINBROWSER_<SOURCE>_<KIND> environment variables. A zero Config omits
// every keyed source from the registry, leaving the keyless default set.
//
// A Config is a secret: its fields must never be logged, and a provider must
// keep them inside request headers (or a private request URL) only — never in a
// returned Result.URL or Result.Detail.
type Config struct {
	// DiscogsToken authenticates the Discogs API
	// (DOMAINBROWSER_DISCOGS_TOKEN).
	DiscogsToken string
	// XBearerToken is the App-Only bearer token for the X/Twitter API
	// (DOMAINBROWSER_X_BEARER_TOKEN).
	XBearerToken string
	// TwitchClientID and TwitchClientSecret authorize the Twitch Helix API
	// (DOMAINBROWSER_TWITCH_CLIENT_ID, DOMAINBROWSER_TWITCH_CLIENT_SECRET).
	TwitchClientID     string
	TwitchClientSecret string
	// CompaniesHouseKey is the HTTP Basic key for the Companies House API
	// (DOMAINBROWSER_COMPANIESHOUSE_KEY).
	CompaniesHouseKey string
	// OpenCorporatesToken authenticates the OpenCorporates API
	// (DOMAINBROWSER_OPENCORPORATES_TOKEN).
	OpenCorporatesToken string
}

// keyedSource describes a source that cannot answer without credentials.
// enabled reports whether cfg supplies what the source needs; new builds the
// provider when it does. A source whose enabled reports false is omitted from
// the registry entirely: it never appears as a row, never receives a request,
// and never reaches the cache. Keyed ports append their sources here.
type keyedSource struct {
	id      string
	enabled func(Config) bool
	new     func(Config) Provider
}

// keyedSources lists the credential-gated sources in registration order, each
// with the predicate that decides whether cfg enables it. A zero Config yields
// the keyless registry unchanged.
var keyedSources = []keyedSource{
	{
		id:      "discogs",
		enabled: func(cfg Config) bool { return cfg.DiscogsToken != "" },
		new:     func(cfg Config) Provider { return NewDiscogs(cfg.DiscogsToken) },
	},
	{
		id:      "x",
		enabled: func(cfg Config) bool { return cfg.XBearerToken != "" },
		new:     func(cfg Config) Provider { return NewX(cfg.XBearerToken) },
	},
	{
		id: "twitch",
		enabled: func(cfg Config) bool {
			return cfg.TwitchClientID != "" && cfg.TwitchClientSecret != ""
		},
		new: func(cfg Config) Provider {
			return NewTwitch(cfg.TwitchClientID, cfg.TwitchClientSecret)
		},
	},
	{
		id:      "companieshouse",
		enabled: func(cfg Config) bool { return cfg.CompaniesHouseKey != "" },
		new:     func(cfg Config) Provider { return NewCompaniesHouse(cfg.CompaniesHouseKey) },
	},
	{
		id:      "opencorporates",
		enabled: func(cfg Config) bool { return cfg.OpenCorporatesToken != "" },
		new:     func(cfg Config) Provider { return NewOpenCorporates(cfg.OpenCorporatesToken) },
	},
}

// KeyedNames returns the ids of the credential-gated sources cfg enables, in
// registration order. Callers use it for startup logging only, so it returns
// names and never touches credential values.
func KeyedNames(cfg Config) []string {
	names := make([]string, 0, len(keyedSources))
	for _, k := range keyedSources {
		if k.enabled(cfg) {
			names = append(names, k.id)
		}
	}
	return names
}

// Default returns a registry populated with the built-in providers. Keyless
// providers are always registered; sources in keyedSources are added only when
// cfg carries the credentials they need, so a zero Config omits them all.
func Default(cfg Config) *Registry {
	r := NewRegistry()
	// Registration cannot fail for the constants below.
	for _, p := range NewDomains() {
		_ = r.Register(p)
	}
	_ = r.Register(NewENS())
	_ = r.Register(NewLemmy())
	_ = r.Register(NewPeerTube())
	_ = r.Register(NewNPM())
	_ = r.Register(NewPyPI())
	_ = r.Register(NewCrates())
	_ = r.Register(NewNuGet())
	_ = r.Register(NewRubyGems())
	_ = r.Register(NewHex())
	_ = r.Register(NewHexUser())
	_ = r.Register(NewPub())
	_ = r.Register(NewHomebrew())
	_ = r.Register(NewHackage())
	_ = r.Register(NewClojars())
	_ = r.Register(NewChocolatey())
	_ = r.Register(NewAUR())
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
	_ = r.Register(NewLastfm())
	_ = r.Register(NewBluesky())
	_ = r.Register(NewMastodon())
	_ = r.Register(NewYouTube())
	_ = r.Register(NewKaggle())
	_ = r.Register(NewDribbble())
	_ = r.Register(NewBehance())
	_ = r.Register(NewReplicate())
	_ = r.Register(NewHackerNews())
	_ = r.Register(NewOpenCollective())
	_ = r.Register(NewVercel())
	_ = r.Register(NewNetlify())
	_ = r.Register(NewSurge())
	_ = r.Register(NewFirebase())
	_ = r.Register(NewNeocities())
	_ = r.Register(NewRender())
	_ = r.Register(NewRailway())
	_ = r.Register(NewHeroku())
	_ = r.Register(NewDenoDeploy())
	_ = r.Register(NewReadTheDocs())
	_ = r.Register(NewEUOrg())
	_ = r.Register(NewDuckDNS())
	_ = r.Register(NewNoIP())
	_ = r.Register(NewDynV6())
	_ = r.Register(NewNotedLol())
	_ = r.Register(NewIsAGoodDev())
	_ = r.Register(NewPagesDev())
	_ = r.Register(NewFly())
	// Keyed sources are omitted entirely when their credentials are absent.
	for _, k := range keyedSources {
		if k.enabled(cfg) {
			_ = r.Register(k.new(cfg))
		}
	}
	return r
}
