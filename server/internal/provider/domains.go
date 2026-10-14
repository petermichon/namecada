package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// BootstrapURL is the IANA RDAP bootstrap file mapping every TLD to the base
// URL of its RDAP registry.
const BootstrapURL = "https://data.iana.org/rdap/dns.json"

// bootstrapTTL is how long a fetched bootstrap map is reused before it is
// refetched, mirroring the frontend's seven-day cache.
const bootstrapTTL = 7 * 24 * time.Hour

// fallbackBases back the TLDs whose registry is stable even when the bootstrap
// file is unreachable, mirroring the frontend FALLBACK map.
var fallbackBases = map[string]string{
	"com": "https://rdap.verisign.com/com/v1/",
	"ai":  "https://rdap.identitydigital.services/rdap/",
	"io":  "https://rdap.identitydigital.services/rdap/",
}

// domainsTLDs are the per-TLD providers registered by Default, in order.
var domainsTLDs = []string{".com", ".org", ".dev", ".io", ".ai", ".app"}

// tldMinLength records the minimum label length for TLDs whose registry rules
// are documented. Google Registry reserves 1-character .dev labels, so no
// 1-character .dev name is registered.
var tldMinLength = map[string]int{".dev": 2}

// reservedLabels are the RFC 6761 special-use labels, reserved in every TLD.
// Some registries (Google's .dev/.app among them) reserve them without creating
// an RDAP record, so a 404 would otherwise read as available.
var reservedLabels = map[string]struct{}{
	"example":   {},
	"test":      {},
	"invalid":   {},
	"localhost": {},
}

// parseBootstrap maps each TLD in the IANA bootstrap document to its registry
// base URL, normalizing the trailing slash. Malformed services are ignored, as
// in the frontend parseBootstrap.
func parseBootstrap(data []byte) map[string]string {
	var doc struct {
		Services []json.RawMessage `json:"services"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return map[string]string{}
	}
	out := make(map[string]string)
	for _, raw := range doc.Services {
		var service []json.RawMessage
		if err := json.Unmarshal(raw, &service); err != nil || len(service) < 2 {
			continue
		}
		var tlds, urls []string
		if json.Unmarshal(service[0], &tlds) != nil ||
			json.Unmarshal(service[1], &urls) != nil || len(urls) == 0 {
			continue
		}
		base := strings.TrimRight(urls[0], "/") + "/"
		for _, tld := range tlds {
			out[strings.ToLower(tld)] = base
		}
	}
	return out
}

// mergeBases returns a new map with extra layered over base (extra wins).
func mergeBases(base, extra map[string]string) map[string]string {
	out := make(map[string]string, len(base)+len(extra))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}

// Bootstrap resolves a TLD to its RDAP registry base URL via the IANA bootstrap
// file, caching the parsed map for TTL. The fallback registries always apply,
// and the bootstrap map wins where both name a TLD. A fetch failure silently
// falls back, mirroring the frontend.
type Bootstrap struct {
	// URL is the bootstrap document; BootstrapURL when defaulted.
	URL string
	// Client is used for the request; DefaultClient when nil.
	Client *http.Client
	// TTL is how long a fetched map is reused; bootstrapTTL when zero.
	TTL time.Duration
	// Now stamps the cache; time.Now when nil.
	Now func() time.Time

	mu     sync.Mutex
	cached map[string]string
	at     time.Time
}

// NewBootstrap returns a resolver bound to the IANA bootstrap URL.
func NewBootstrap() *Bootstrap {
	return &Bootstrap{URL: BootstrapURL, Client: DefaultClient, TTL: bootstrapTTL, Now: time.Now}
}

// sharedBootstrap is the resolver the registered domain providers share, so its
// cache is loaded once per process.
var sharedBootstrap = NewBootstrap()

func (b *Bootstrap) ttl() time.Duration {
	if b.TTL > 0 {
		return b.TTL
	}
	return bootstrapTTL
}

func (b *Bootstrap) now() time.Time {
	if b.Now != nil {
		return b.Now()
	}
	return time.Now()
}

// Resolve returns the RDAP base URL for tld (with or without a leading dot) and
// whether one is known.
func (b *Bootstrap) Resolve(ctx context.Context, tld string) (string, bool) {
	base, ok := b.load(ctx)[strings.ToLower(strings.TrimPrefix(tld, "."))]
	return base, ok
}

func (b *Bootstrap) load(ctx context.Context) map[string]string {
	b.mu.Lock()
	cached, at := b.cached, b.at
	b.mu.Unlock()

	if cached != nil && b.now().Sub(at) < b.ttl() {
		return cached
	}

	merged := mergeBases(fallbackBases, b.fetch(ctx))
	b.mu.Lock()
	b.cached = merged
	b.at = b.now()
	b.mu.Unlock()
	return merged
}

// fetch retrieves and parses the bootstrap document. Any failure yields nil so
// the caller falls back to the static registries.
func (b *Bootstrap) fetch(ctx context.Context) map[string]string {
	target := b.URL
	if target == "" {
		target = BootstrapURL
	}
	client := b.Client
	if client == nil {
		client = DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer drainAndClose(resp)
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil
	}
	return parseBootstrap(data)
}

// Domain checks whether a label is registered under one TLD using RDAP. Each
// registered TLD is its own provider, sharing the "domains" platform.
type Domain struct {
	// TLD is the TLD including its leading dot, e.g. ".dev".
	TLD string
	// BaseURL overrides the RDAP registry base; when empty the base is resolved
	// from the bootstrap (or fallback) registries.
	BaseURL string
	// Bootstrap resolves the registry base when BaseURL is empty; the shared
	// resolver is used when nil.
	Bootstrap *Bootstrap
	// Client is used for the request; DefaultClient when nil.
	Client *http.Client
	// Min is the shortest label this TLD accepts; 0 means no minimum.
	Min int
}

// NewDomain returns a provider for tld (with leading dot) bound to the shared
// bootstrap resolver.
func NewDomain(tld string) *Domain {
	return &Domain{TLD: tld, Bootstrap: sharedBootstrap, Client: DefaultClient, Min: tldMinLength[tld]}
}

// NewDomains returns the per-TLD RDAP domain providers in order.
func NewDomains() []Provider {
	out := make([]Provider, 0, len(domainsTLDs))
	for _, tld := range domainsTLDs {
		out = append(out, NewDomain(tld))
	}
	return out
}

func (d *Domain) ID() string       { return "domain" + d.TLD }
func (d *Domain) Platform() string { return "domains" }

func (d *Domain) Name(label string) string { return label + d.TLD }

func (d *Domain) URL(label string) string { return "https://" + label + d.TLD }

// MinLength reports the TLD's minimum label length (2 for .dev, 0 otherwise).
func (d *Domain) MinLength() int { return d.Min }

// Check reports a reserved label as taken without a request; otherwise it looks
// up the TLD's registry and reads the RDAP response status: 200 = taken,
// 404 = available, anything else (or no registry) = unknown.
func (d *Domain) Check(ctx context.Context, label string) (Result, error) {
	if _, ok := reservedLabels[strings.ToLower(label)]; ok {
		return Result{Status: StatusTaken}, nil
	}

	base := d.BaseURL
	if base == "" {
		resolver := d.Bootstrap
		if resolver == nil {
			resolver = sharedBootstrap
		}
		got, ok := resolver.Resolve(ctx, d.TLD)
		if !ok {
			return Result{Status: StatusUnknown, Detail: "no RDAP endpoint"}, nil
		}
		base = got
	}

	target := base + "domain/" + url.PathEscape(label) + d.TLD
	return checkHTTP(ctx, d.Client, target, defaultCodes)
}
