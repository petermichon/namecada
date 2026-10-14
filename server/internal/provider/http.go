package provider

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

// DefaultTimeout is the request timeout on DefaultClient. It bounds a single
// source's HTTP call so a hung server cannot stall a scan indefinitely, which
// the timeout-less http.DefaultClient would allow.
const DefaultTimeout = 10 * time.Second

// DefaultClient is the request client providers use when none is supplied. Its
// Timeout applies to the whole request, including reading the body.
var DefaultClient = &http.Client{Timeout: DefaultTimeout}

// UserAgent identifies the client on every outbound request. Several sources
// (crates.io, Wikimedia, Chess.com, Open Collective) reject the Go default UA
// with a 403, so the descriptive product string below is set instead.
const UserAgent = "domain-browser/1.0 (+https://github.com/petermichon/domain-browser)"

// defaultAccept is the Accept header providers send unless they override it.
// Most sources only speak JSON, but some (e.g. Mastodon's profile page) reject
// an application/json Accept on an HTML route.
const defaultAccept = "application/json"

// htmlAccept is the Accept header for providers whose probe hits an HTML route.
// It prefers HTML but allows anything, so a strictly-negotiating app that would
// answer 406 to application/json (e.g. a claimed Railway app) still responds.
const htmlAccept = "text/html,*/*"

// setCommonHeaders applies the headers every provider request carries. An empty
// accept falls back to defaultAccept.
func setCommonHeaders(req *http.Request, accept string) {
	if accept == "" {
		accept = defaultAccept
	}
	req.Header.Set("Accept", accept)
	req.Header.Set("User-Agent", UserAgent)
}

// codeMap maps HTTP status codes to a check outcome.
type codeMap struct {
	available []int
	taken     []int
}

// defaultCodes treats 404 as free and 200 as taken, matching the frontend
// checkHttp defaults.
var defaultCodes = codeMap{available: []int{http.StatusNotFound}, taken: []int{http.StatusOK}}

// doGet performs a GET with the shared JSON Accept header; a nil client falls
// back to DefaultClient so every request carries a timeout.
func doGet(ctx context.Context, client *http.Client, url string) (*http.Response, error) {
	return doGetAccept(ctx, client, url, defaultAccept)
}

// doGetAccept is doGet with an Accept override; an empty accept keeps
// defaultAccept. It exists for sources that serve HTML rather than JSON.
func doGetAccept(ctx context.Context, client *http.Client, url, accept string) (*http.Response, error) {
	if client == nil {
		client = DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	setCommonHeaders(req, accept)
	return client.Do(req)
}

// doPost performs a POST with the shared JSON Accept header, a request body and
// an explicit Content-Type; a nil client falls back to DefaultClient so every
// request carries a timeout. It exists for the GraphQL sources, whose verdict
// comes from a query body rather than the request path.
func doPost(ctx context.Context, client *http.Client, url, contentType string, body io.Reader) (*http.Response, error) {
	if client == nil {
		client = DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	setCommonHeaders(req, defaultAccept)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return client.Do(req)
}

// doGetHeaders performs a GET with the shared headers plus extra request
// headers. It exists for the credential-gated sources, which carry a token (and
// sometimes an identity header) on every request. The extra values are secrets:
// callers must keep them out of a returned Result, a log, and a URL.
func doGetHeaders(ctx context.Context, client *http.Client, url string, headers map[string]string) (*http.Response, error) {
	if client == nil {
		client = DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	setCommonHeaders(req, defaultAccept)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return client.Do(req)
}

// drainAndClose fully reads and closes a response body so the connection can be
// reused.
func drainAndClose(resp *http.Response) {
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}

// checkHTTP performs a GET and maps the response status through codes. Any
// status that is neither taken nor available yields StatusUnknown with no
// error; transport failures are returned as errors.
func checkHTTP(ctx context.Context, client *http.Client, url string, codes codeMap) (Result, error) {
	return checkHTTPAccept(ctx, client, url, codes, defaultAccept)
}

// checkHTTPAccept is checkHTTP with an Accept override; an empty accept keeps
// defaultAccept.
func checkHTTPAccept(ctx context.Context, client *http.Client, url string, codes codeMap, accept string) (Result, error) {
	resp, err := doGetAccept(ctx, client, url, accept)
	if err != nil {
		return Result{}, err
	}
	defer drainAndClose(resp)

	result := codes.classify(resp.StatusCode)
	// A bot-challenge interstitial can answer 200 (Cloudflare's "Just a moment"
	// / "Client Challenge" pages), so a status-only source would otherwise
	// report every label as taken. Only a would-be taken verdict is inspected,
	// and never cached as unknown, so a real taken response is untouched.
	if result.Status == StatusTaken && isChallenge(resp) {
		return Result{Status: StatusUnknown}, nil
	}
	return result, nil
}

// challengeMarkers are substrings of the HTML bot-challenge interstitials that
// servers such as Cloudflare return with a 200 status. They are matched
// case-insensitively against a bounded body prefix.
var challengeMarkers = []string{"client challenge", "just a moment"}

// challengeBodyLimit bounds how much of a taken response body is scanned for
// challenge markers, keeping the extra read cheap.
const challengeBodyLimit = 8 << 10

// isChallenge reports whether resp looks like a bot-challenge page rather than a
// real taken answer. It checks Cloudflare's cf-mitigated header and a bounded
// prefix of the body. Callers invoke it only on the taken path.
func isChallenge(resp *http.Response) bool {
	if strings.EqualFold(resp.Header.Get("cf-mitigated"), "challenge") {
		return true
	}
	buf := make([]byte, challengeBodyLimit)
	n, err := io.ReadFull(resp.Body, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return false
	}
	body := strings.ToLower(string(buf[:n]))
	for _, marker := range challengeMarkers {
		if strings.Contains(body, marker) {
			return true
		}
	}
	return false
}

func (c codeMap) classify(status int) Result {
	for _, code := range c.taken {
		if status == code {
			return Result{Status: StatusTaken}
		}
	}
	for _, code := range c.available {
		if status == code {
			return Result{Status: StatusAvailable}
		}
	}
	return Result{Status: StatusUnknown}
}
