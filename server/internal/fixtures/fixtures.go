// Package fixtures captures the current source inventory and raw HTTP
// transcripts. The artifacts are a parity oracle for the engine rewrite: the
// Rust implementation can replay the recorded responses and compare its verdict
// with the Go result captured here.
package fixtures

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/petermichon/domain-browser/server/internal/provider"
)

// DefaultBodyLimit is how much of a request or response body is written to a
// fixture. The provider always sees the full body; only the snapshot is capped.
const DefaultBodyLimit = 256 << 10

// Source is one row of the source inventory, for a single label.
type Source struct {
	ID           string `json:"id"`
	Platform     string `json:"platform"`
	PlatformName string `json:"platformName"`
	Name         string `json:"name"`
	NameTemplate string `json:"nameTemplate"`
	URL          string `json:"url"`
	URLTemplate  string `json:"urlTemplate"`
	Paid         bool   `json:"paid"`
	MinLength    int    `json:"minLength"`
	MaxLength    int    `json:"maxLength"`
	Keyed        bool   `json:"keyed"`
}

// Inventory is the complete source set: the keyless providers a zero Config
// registers plus every credential-gated provider. It performs no requests.
type Inventory struct {
	GeneratedAt  string   `json:"generatedAt"`
	Label        string   `json:"label"`
	SourceCount  int      `json:"sourceCount"`
	EnabledCount int      `json:"enabledCount"`
	KeyedCount   int      `json:"keyedCount"`
	Sources      []Source `json:"sources"`
}

// Request is one outbound HTTP request and its response (or transport error).
type Request struct {
	Method        string            `json:"method"`
	URL           string            `json:"url"`
	Headers       map[string]string `json:"headers,omitempty"`
	Body          string            `json:"body,omitempty"`
	BodyEncoding  string            `json:"bodyEncoding,omitempty"`
	BodyTruncated bool              `json:"bodyTruncated,omitempty"`
	Response      *Response         `json:"response,omitempty"`
	Error         string            `json:"error,omitempty"`
}

// Response is the captured part of an HTTP response.
type Response struct {
	Status        int               `json:"status"`
	Headers       map[string]string `json:"headers,omitempty"`
	Body          string            `json:"body,omitempty"`
	BodyEncoding  string            `json:"bodyEncoding,omitempty"`
	BodyTruncated bool              `json:"bodyTruncated,omitempty"`
}

// Fixture is one provider's result for a label plus the requests it made.
type Fixture struct {
	Source       string          `json:"source"`
	Platform     string          `json:"platform"`
	PlatformName string          `json:"platformName"`
	Label        string          `json:"label"`
	Name         string          `json:"name"`
	URL          string          `json:"url"`
	Paid         bool            `json:"paid"`
	Status       provider.Status `json:"status,omitempty"`
	Detail       string          `json:"detail,omitempty"`
	Error        string          `json:"error,omitempty"`
	DurationMS   int64           `json:"durationMs"`
	Requests     []Request       `json:"requests"`
}

// ExportOptions controls an export run.
type ExportOptions struct {
	// OutDir receives sources.json and one directory per label.
	OutDir string
	// Labels are captured in order. No labels means an inventory-only export.
	Labels []string
	// Concurrency bounds simultaneous provider checks; 0 uses 8.
	Concurrency int
	// BodyLimit caps captured bodies; 0 uses DefaultBodyLimit.
	BodyLimit int
}

// LabelSummary describes one captured label.
type LabelSummary struct {
	Label    string
	Dir      string
	Fixtures int
	Errors   int
}

// ExportSummary describes a finished export.
type ExportSummary struct {
	SourcesPath  string
	SourceCount  int
	EnabledCount int
	KeyedCount   int
	Labels       []LabelSummary
}

// Export writes the inventory and, for every label, one fixture per provider.
// It swaps provider.DefaultClient for a recording client so the registered
// providers capture their real outbound requests; the original client is
// restored before Export returns.
func Export(ctx context.Context, opt ExportOptions) (ExportSummary, error) {
	if opt.OutDir == "" {
		return ExportSummary{}, fmt.Errorf("fixtures: output directory is required")
	}
	concurrency := opt.Concurrency
	if concurrency < 1 {
		concurrency = 8
	}
	bodyLimit := opt.BodyLimit
	if bodyLimit < 1 {
		bodyLimit = DefaultBodyLimit
	}

	rec := NewRecorder(http.DefaultTransport, bodyLimit)
	recClient := &http.Client{Transport: rec, Timeout: provider.DefaultTimeout}
	oldClient := provider.DefaultClient
	provider.DefaultClient = recClient
	defer func() { provider.DefaultClient = oldClient }()

	keyless := provider.Default(provider.Config{}).All()
	keyed := provider.KeyedProviders()

	label := "example"
	if len(opt.Labels) > 0 {
		label = opt.Labels[0]
	}
	inv := BuildInventory(label, keyless, keyed)
	sourcesPath, err := WriteInventory(opt.OutDir, inv)
	if err != nil {
		return ExportSummary{}, err
	}

	summary := ExportSummary{
		SourcesPath:  sourcesPath,
		SourceCount:  inv.SourceCount,
		EnabledCount: inv.EnabledCount,
		KeyedCount:   inv.KeyedCount,
	}
	for _, label := range opt.Labels {
		captured := CaptureAll(ctx, keyless, label, rec, concurrency)
		dir, err := WriteFixtures(opt.OutDir, label, captured)
		if err != nil {
			return summary, err
		}
		ls := LabelSummary{Label: label, Dir: dir, Fixtures: len(captured)}
		for _, f := range captured {
			if f.Error != "" {
				ls.Errors++
			}
		}
		summary.Labels = append(summary.Labels, ls)
	}
	return summary, nil
}

// BuildInventory describes keyless and keyed providers for label. It never
// touches the network or logs a credential.
func BuildInventory(label string, keyless, keyed []provider.Provider) Inventory {
	sources := make([]Source, 0, len(keyless)+len(keyed))
	for _, p := range keyless {
		sources = append(sources, describe(p, label, false))
	}
	for _, p := range keyed {
		sources = append(sources, describe(p, label, true))
	}
	return Inventory{
		GeneratedAt:  time.Now().UTC().Format(time.RFC3339),
		Label:        label,
		SourceCount:  len(sources),
		EnabledCount: len(keyless),
		KeyedCount:   len(keyed),
		Sources:      sources,
	}
}

func describe(p provider.Provider, label string, keyed bool) Source {
	return Source{
		ID:           p.ID(),
		Platform:     p.Platform(),
		PlatformName: provider.PlatformName(p.Platform()),
		Name:         p.Name(label),
		NameTemplate: provider.NameTemplate(p),
		URL:          p.URL(label),
		URLTemplate:  provider.URLTemplate(p),
		Paid:         provider.Paid(p),
		MinLength:    provider.MinLength(p),
		MaxLength:    provider.MaxLength(p),
		Keyed:        keyed,
	}
}

// WriteInventory writes sources.json under outDir, creating the directory.
func WriteInventory(outDir string, inv Inventory) (string, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", fmt.Errorf("fixtures: create %s: %w", outDir, err)
	}
	path := filepath.Join(outDir, "sources.json")
	return path, writeJSON(path, inv)
}

// CaptureAll checks every provider against label concurrently and returns the
// fixtures in provider order.
func CaptureAll(ctx context.Context, providers []provider.Provider, label string, rec *Recorder, concurrency int) []Fixture {
	out := make([]Fixture, len(providers))
	if len(providers) == 0 {
		return out
	}
	if concurrency < 1 {
		concurrency = 8
	}
	if concurrency > len(providers) {
		concurrency = len(providers)
	}

	jobs := make(chan int)
	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for idx := range jobs {
				out[idx] = Capture(ctx, providers[idx], label, rec)
			}
		}()
	}

	go func() {
		defer close(jobs)
		for i := range providers {
			select {
			case jobs <- i:
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Wait()
	return out
}

// Capture runs one provider with a context tagged with its id, then takes the
// requests the recorder saw for that id. The provider's own error is preserved
// separately from its verdict.
func Capture(ctx context.Context, p provider.Provider, label string, rec *Recorder) Fixture {
	start := time.Now()
	res, err := p.Check(withSource(ctx, p.ID()), label)

	f := Fixture{
		Source:       p.ID(),
		Platform:     p.Platform(),
		PlatformName: provider.PlatformName(p.Platform()),
		Label:        label,
		Name:         p.Name(label),
		URL:          p.URL(label),
		Paid:         provider.Paid(p),
		DurationMS:   time.Since(start).Milliseconds(),
		Requests:     rec.Take(p.ID()),
	}
	if f.Requests == nil {
		f.Requests = []Request{}
	}
	if err != nil {
		f.Error = err.Error()
		return f
	}
	if res.URL != "" {
		f.URL = res.URL
	}
	f.Status = res.Status
	f.Detail = res.Detail
	return f
}

// WriteFixtures writes one <source>.json per fixture under outDir/label.
func WriteFixtures(outDir, label string, captured []Fixture) (string, error) {
	dir := filepath.Join(outDir, label)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("fixtures: create %s: %w", dir, err)
	}
	for _, f := range captured {
		if f.Source == "" || strings.ContainsAny(f.Source, `/\`) {
			return "", fmt.Errorf("fixtures: invalid source id %q", f.Source)
		}
		path := filepath.Join(dir, f.Source+".json")
		if err := writeJSON(path, f); err != nil {
			return "", err
		}
	}
	return dir, nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("fixtures: encode %s: %w", path, err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("fixtures: write %s: %w", path, err)
	}
	return nil
}

// sourceKey tags an outbound request with the provider that made it.
type sourceKey struct{}

func withSource(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, sourceKey{}, id)
}

// Recorder is an http.RoundTripper that snapshots requests and responses by the
// provider id carried in the request context.
type Recorder struct {
	base      http.RoundTripper
	bodyLimit int

	mu       sync.Mutex
	bySource map[string][]Request
}

// NewRecorder wraps base (http.DefaultTransport when nil).
func NewRecorder(base http.RoundTripper, bodyLimit int) *Recorder {
	if base == nil {
		base = http.DefaultTransport
	}
	if bodyLimit < 1 {
		bodyLimit = DefaultBodyLimit
	}
	return &Recorder{base: base, bodyLimit: bodyLimit, bySource: make(map[string][]Request)}
}

// RoundTrip forwards req and records a snapshot. The body handed to the caller
// is always the full response, even when the recorded snapshot is truncated.
func (r *Recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	source, _ := req.Context().Value(sourceKey{}).(string)
	if source == "" {
		source = "unknown"
	}

	rec := Request{
		Method:  req.Method,
		URL:     redactURL(req.URL.String()),
		Headers: requestHeaders(req.Header),
	}
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		_ = req.Body.Close()
		req.Body = io.NopCloser(bytes.NewReader(body))
		if err != nil {
			rec.Error = "read request body: " + err.Error()
		}
		prefix, truncated := limitBody(body, r.bodyLimit)
		rec.Body, rec.BodyEncoding = encodeBody(prefix)
		rec.BodyTruncated = truncated
	}

	resp, err := r.base.RoundTrip(req)
	if err != nil {
		rec.Error = err.Error()
		r.record(source, rec)
		return nil, err
	}

	body, readErr := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(body))
	if readErr != nil {
		rec.Error = "read response body: " + readErr.Error()
	}

	prefix, truncated := limitBody(body, r.bodyLimit)
	rec.Response = &Response{
		Status:        resp.StatusCode,
		Headers:       responseHeaders(resp.Header),
		Body:          "",
		BodyEncoding:  "",
		BodyTruncated: truncated,
	}
	rec.Response.Body, rec.Response.BodyEncoding = encodeBody(prefix)
	r.record(source, rec)
	return resp, nil
}

func (r *Recorder) record(source string, rec Request) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bySource[source] = append(r.bySource[source], rec)
}

// Take returns and clears the requests recorded for source.
func (r *Recorder) Take(source string) []Request {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := r.bySource[source]
	delete(r.bySource, source)
	return out
}

func limitBody(body []byte, limit int) ([]byte, bool) {
	if len(body) <= limit {
		return body, false
	}
	return body[:limit], true
}

func encodeBody(body []byte) (string, string) {
	if len(body) == 0 {
		return "", ""
	}
	if utf8.Valid(body) {
		return string(body), "utf8"
	}
	return base64.StdEncoding.EncodeToString(body), "base64"
}

var secretHeaders = map[string]bool{
	"authorization": true,
	"cookie":        true,
	"x-api-key":     true,
	"api-key":       true,
	"x-auth-token":  true,
	"client-secret": true,
}

func requestHeaders(h http.Header) map[string]string {
	out := make(map[string]string, len(h))
	for key, values := range h {
		if secretHeaders[strings.ToLower(key)] {
			out[key] = "[REDACTED]"
			continue
		}
		out[key] = strings.Join(values, ", ")
	}
	return out
}

var responseHeaderAllowlist = map[string]bool{
	"content-type":     true,
	"content-language": true,
	"location":         true,
	"cf-mitigated":     true,
	"retry-after":      true,
}

func responseHeaders(h http.Header) map[string]string {
	out := make(map[string]string)
	for key, values := range h {
		if !responseHeaderAllowlist[strings.ToLower(key)] {
			continue
		}
		out[key] = strings.Join(values, ", ")
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

var secretQueryKeys = map[string]bool{
	"token":         true,
	"access_token":  true,
	"apikey":        true,
	"api_key":       true,
	"key":           true,
	"client_secret": true,
	"password":      true,
}

func redactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	q := u.Query()
	changed := false
	for key := range q {
		if secretQueryKeys[strings.ToLower(key)] {
			q.Set(key, "[REDACTED]")
			changed = true
		}
	}
	if !changed {
		return raw
	}
	u.RawQuery = q.Encode()
	return u.String()
}
