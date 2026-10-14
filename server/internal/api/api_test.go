package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/petermichon/domain-browser/server/internal/engine"
	"github.com/petermichon/domain-browser/server/internal/provider"
	"github.com/petermichon/domain-browser/server/internal/store"
)

// fake is a provider with a fixed result and an optional call counter.
type fake struct {
	id     string
	result provider.Result
	calls  *atomic.Int32
}

func (f *fake) ID() string               { return f.id }
func (f *fake) Platform() string         { return f.id }
func (f *fake) Name(label string) string { return f.id + ":" + label }
func (f *fake) URL(label string) string  { return "https://" + f.id + ".test/" + label }
func (f *fake) Check(context.Context, string) (provider.Result, error) {
	if f.calls != nil {
		f.calls.Add(1)
	}
	return f.result, nil
}

// newTestServer builds a Server over the given providers and returns a test
// HTTP server plus the engine's memory store (shared across requests).
func newTestServer(t *testing.T, origin string, ps ...provider.Provider) *httptest.Server {
	t.Helper()
	eng := engine.NewWithStore(store.NewMemory(), ps...)
	return httptest.NewServer(New(eng, origin, nil).Handler())
}

func TestCheckStreamsNDJSON(t *testing.T) {
	srv := newTestServer(t, DefaultOrigin,
		&fake{id: "a", result: provider.Result{Status: provider.StatusAvailable, Detail: "free"}},
		&fake{id: "b", result: provider.Result{Status: provider.StatusTaken}},
	)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/check?name=acema")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/x-ndjson" {
		t.Errorf("Content-Type = %q, want application/x-ndjson", ct)
	}

	byID := map[string]engine.Result{}
	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		var res engine.Result
		if err := json.Unmarshal(sc.Bytes(), &res); err != nil {
			t.Fatalf("decode %q: %v", sc.Text(), err)
		}
		byID[res.ID] = res
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}

	if len(byID) != 2 {
		t.Fatalf("got %d results, want 2: %+v", len(byID), byID)
	}
	if r := byID["a"]; r.Status != provider.StatusAvailable || r.Detail != "free" || r.URL != "https://a.test/acema" {
		t.Errorf("a = %+v", r)
	}
	if r := byID["b"]; r.Status != provider.StatusTaken {
		t.Errorf("b = %+v", r)
	}
}

// TestCheckLogsRequest pins the structured request log: one line per request
// with the method, label, duration, and streamed result count.
func TestCheckLogsRequest(t *testing.T) {
	eng := engine.NewWithStore(store.NewMemory(),
		&fake{id: "a", result: provider.Result{Status: provider.StatusAvailable}},
		&fake{id: "b", result: provider.Result{Status: provider.StatusTaken}},
	)
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	srv := httptest.NewServer(New(eng, DefaultOrigin, logger).Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/check?name=acema")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	logged := buf.String()
	for _, want := range []string{"msg=request", "method=GET", "label=acema", "duration=", "results=2"} {
		if !strings.Contains(logged, want) {
			t.Errorf("log = %q, want to contain %q", logged, want)
		}
	}
}

func TestCheckAllowsAppOrigin(t *testing.T) {
	srv := newTestServer(t, "http://localhost:5173", &fake{id: "a"})
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/check?name=acema", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("Access-Control-Allow-Origin = %q, want http://localhost:5173", got)
	}
}

func TestCheckPreflight(t *testing.T) {
	srv := newTestServer(t, "http://localhost:5173", &fake{id: "a"})
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodOptions, srv.URL+"/check?name=acema", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("Access-Control-Allow-Origin = %q", got)
	}
	if got := resp.Header.Get("Access-Control-Allow-Methods"); !strings.Contains(got, "GET") {
		t.Errorf("Access-Control-Allow-Methods = %q, want GET", got)
	}
}

func TestCheckRequiresName(t *testing.T) {
	srv := newTestServer(t, DefaultOrigin, &fake{id: "a"})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/check")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if body["error"] == "" {
		t.Errorf("error body = %+v, want a message", body)
	}
}

func TestNormalizeLabel(t *testing.T) {
	tests := []struct {
		in   string
		want string
		ok   bool
	}{
		{"  AceMa ", "acema", true},
		{"a-1-b", "a-1-b", true},
		{"a", "a", true},
		{"A", "a", true},
		{"acema.com", "", false},
		{"www.bbc.co.uk", "", false},
		{"", "", false},
		{"   ", "", false},
		{"a b", "", false},
		{"-abc", "", false},
		{"abc-", "", false},
		{strings.Repeat("a", 63), strings.Repeat("a", 63), true},
		{strings.Repeat("a", 64), "", false},
	}
	for _, tc := range tests {
		got, ok := normalizeLabel(tc.in)
		if got != tc.want || ok != tc.ok {
			t.Errorf("normalizeLabel(%q) = (%q, %v), want (%q, %v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestCheckNormalizesName(t *testing.T) {
	srv := newTestServer(t, DefaultOrigin, &fake{id: "a"})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/check?name=" + url.QueryEscape("  AceMa "))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var res engine.Result
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if res.URL != "https://a.test/acema" {
		t.Errorf("URL = %q, want https://a.test/acema (trimmed and lowercased)", res.URL)
	}
}

func TestCheckRejectsInvalidName(t *testing.T) {
	srv := newTestServer(t, DefaultOrigin, &fake{id: "a"})
	defer srv.Close()

	for _, name := range []string{"acema.com", "-abc", "abc-", "a b"} {
		resp, err := http.Get(srv.URL + "/check?name=" + url.QueryEscape(name))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			resp.Body.Close()
			t.Fatalf("name %q: status = %d, want 400", name, resp.StatusCode)
		}
		var body map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			resp.Body.Close()
			t.Fatalf("name %q: decode error body: %v", name, err)
		}
		resp.Body.Close()
		if body["error"] == "" {
			t.Errorf("name %q: error body = %+v, want a message", name, body)
		}
	}
}

func TestCheckRejectsNonGet(t *testing.T) {
	srv := newTestServer(t, DefaultOrigin, &fake{id: "a"})
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/check?name=acema", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", resp.StatusCode)
	}
	if allow := resp.Header.Get("Allow"); !strings.Contains(allow, "GET") {
		t.Errorf("Allow = %q, want GET", allow)
	}
}

func TestCheckReusesMemoryStoreAcrossRequests(t *testing.T) {
	var calls atomic.Int32
	srv := newTestServer(t, DefaultOrigin,
		&fake{id: "a", result: provider.Result{Status: provider.StatusTaken, Detail: "found"}, calls: &calls},
	)
	defer srv.Close()

	for i := 0; i < 2; i++ {
		resp, err := http.Get(srv.URL + "/check?name=acema")
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
	}
	if got := calls.Load(); got != 1 {
		t.Errorf("provider calls = %d, want 1 (second request served from the memory store)", got)
	}
}
