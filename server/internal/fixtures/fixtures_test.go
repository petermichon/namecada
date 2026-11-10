package fixtures

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/petermichon/domain-browser/server/internal/provider"
)

// TestRecorderCapturesAndRestores proves the snapshot keeps the request URL,
// headers and body, records the response, and leaves the full body readable by
// the provider.
func TestRecorderCapturesAndRestores(t *testing.T) {
	const responseBody = `{"name":"acema"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "secret" {
			t.Errorf("Authorization = %q, want secret", got)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"query":"ok"}` {
			t.Errorf("request body = %q", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Set-Cookie", "session=secret")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(responseBody))
	}))
	defer srv.Close()

	rec := NewRecorder(srv.Client().Transport, DefaultBodyLimit)
	client := &http.Client{Transport: rec}
	req, err := http.NewRequestWithContext(
		withSource(context.Background(), "stub"),
		http.MethodPost,
		srv.URL+"/check?token=secret&name=acema",
		strings.NewReader(`{"query":"ok"}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "secret")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do: %v", err)
	}
	defer resp.Body.Close()
	gotBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if string(gotBody) != responseBody {
		t.Errorf("provider body = %q, want %q", gotBody, responseBody)
	}

	records := rec.Take("stub")
	if len(records) != 1 {
		t.Fatalf("records = %d, want 1", len(records))
	}
	got := records[0]
	if got.Method != http.MethodPost {
		t.Errorf("method = %q", got.Method)
	}
	if got.URL != srv.URL+"/check?name=acema&token=%5BREDACTED%5D" {
		t.Errorf("url = %q", got.URL)
	}
	if got.Headers["Authorization"] != "[REDACTED]" {
		t.Errorf("authorization header = %q", got.Headers["Authorization"])
	}
	if got.Body != `{"query":"ok"}` || got.BodyEncoding != "utf8" {
		t.Errorf("request body = %q (%q)", got.Body, got.BodyEncoding)
	}
	if got.Response == nil || got.Response.Status != http.StatusCreated {
		t.Fatalf("response = %+v", got.Response)
	}
	if got.Response.Body != responseBody {
		t.Errorf("response body = %q", got.Response.Body)
	}
	if got.Response.Headers["Content-Type"] != "application/json" {
		t.Errorf("response content-type = %q", got.Response.Headers["Content-Type"])
	}
	if _, ok := got.Response.Headers["Set-Cookie"]; ok {
		t.Error("Set-Cookie was recorded")
	}
}

// TestRecorderTruncatesSnapshotNotProviderBody proves the recorded prefix can be
// shorter than the body the provider reads.
func TestRecorderTruncatesSnapshotNotProviderBody(t *testing.T) {
	const responseBody = "0123456789"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(responseBody))
	}))
	defer srv.Close()

	rec := NewRecorder(srv.Client().Transport, 4)
	client := &http.Client{Transport: rec}
	req, _ := http.NewRequestWithContext(withSource(context.Background(), "stub"), http.MethodGet, srv.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	got, _ := io.ReadAll(resp.Body)
	if string(got) != responseBody {
		t.Fatalf("provider body = %q, want full body", got)
	}

	records := rec.Take("stub")
	if len(records) != 1 || records[0].Response == nil {
		t.Fatalf("records = %+v", records)
	}
	if records[0].Response.Body != "0123" {
		t.Errorf("snapshot body = %q, want 0123", records[0].Response.Body)
	}
	if !records[0].Response.BodyTruncated {
		t.Error("snapshot was not marked truncated")
	}
}

// TestCaptureKeepsProviderErrorSeparateFromVerdict pins that a transport error
// is recorded as an error and not silently turned into a status.
func TestCaptureKeepsProviderErrorSeparateFromVerdict(t *testing.T) {
	rec := NewRecorder(nil, DefaultBodyLimit)
	p := stubProvider{
		id:     "stub",
		url:    "http://127.0.0.1:1/check",
		client: &http.Client{Transport: rec},
	}
	f := Capture(context.Background(), p, "acema", rec)
	if f.Error == "" {
		t.Fatal("expected a transport error")
	}
	if f.Status != "" {
		t.Errorf("status = %q, want empty on transport error", f.Status)
	}
	if len(f.Requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(f.Requests))
	}
}

func TestBuildInventory(t *testing.T) {
	keyless := []provider.Provider{provider.NewNPM(), provider.NewDomain(".dev")}
	keyed := []provider.Provider{provider.NewDiscogs("placeholder")}

	inv := BuildInventory("acema", keyless, keyed)
	if inv.SourceCount != 3 || inv.EnabledCount != 2 || inv.KeyedCount != 1 {
		t.Fatalf("inventory counts = %+v", inv)
	}
	if inv.Sources[0].ID != "npm" || inv.Sources[0].Keyed {
		t.Errorf("first source = %+v", inv.Sources[0])
	}
	if inv.Sources[0].NameTemplate != "%s" || inv.Sources[0].URLTemplate != "https://www.npmjs.com/package/%s" {
		t.Errorf("npm templates = %q, %q", inv.Sources[0].NameTemplate, inv.Sources[0].URLTemplate)
	}
	if inv.Sources[1].MinLength != 2 {
		t.Errorf("domain min length = %d, want 2", inv.Sources[1].MinLength)
	}
	if inv.Sources[1].NameTemplate != "%s.dev" || inv.Sources[1].URLTemplate != "https://%s.dev" {
		t.Errorf("domain templates = %q, %q", inv.Sources[1].NameTemplate, inv.Sources[1].URLTemplate)
	}
	if last := inv.Sources[2]; last.ID != "discogs" || !last.Keyed {
		t.Errorf("keyed source = %+v", last)
	}
}

func TestWriteInventoryAndFixtures(t *testing.T) {
	dir := t.TempDir()
	inv := BuildInventory("acema", []provider.Provider{provider.NewNPM()}, nil)
	path, err := WriteInventory(dir, inv)
	if err != nil {
		t.Fatalf("WriteInventory: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got Inventory
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("inventory json: %v", err)
	}
	if got.SourceCount != 1 || got.Sources[0].ID != "npm" {
		t.Fatalf("inventory = %+v", got)
	}

	fixtures := []Fixture{{
		Source:   "npm",
		Label:    "acema",
		Status:   provider.StatusAvailable,
		Requests: []Request{},
	}}
	outDir, err := WriteFixtures(dir, "acema", fixtures)
	if err != nil {
		t.Fatalf("WriteFixtures: %v", err)
	}
	if outDir != filepath.Join(dir, "acema") {
		t.Errorf("outDir = %q", outDir)
	}
	if _, err := os.Stat(filepath.Join(outDir, "npm.json")); err != nil {
		t.Errorf("fixture missing: %v", err)
	}
}

type stubProvider struct {
	id     string
	url    string
	client *http.Client
}

func (s stubProvider) ID() string               { return s.id }
func (s stubProvider) Platform() string         { return "stub" }
func (s stubProvider) Name(label string) string { return label }
func (s stubProvider) URL(label string) string  { return s.url }

func (s stubProvider) Check(ctx context.Context, _ string) (provider.Result, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.url, nil)
	if err != nil {
		return provider.Result{}, err
	}
	client := s.client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return provider.Result{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return provider.Result{Status: provider.StatusTaken}, nil
	}
	return provider.Result{Status: provider.StatusAvailable}, nil
}
