package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/petermichon/domain-browser/server/internal/engine"
	"github.com/petermichon/domain-browser/server/internal/provider"

	_ "modernc.org/sqlite"
)

// fakeProvider is a deterministic provider for the CLI tests.
type fakeProvider struct {
	id     string
	status provider.Status
	detail string
}

func (f fakeProvider) ID() string               { return f.id }
func (f fakeProvider) Platform() string         { return "p-" + f.id }
func (f fakeProvider) Name(label string) string { return f.id + ":" + label }
func (f fakeProvider) URL(label string) string  { return "https://" + f.id + ".test/" + label }
func (f fakeProvider) Check(context.Context, string) (provider.Result, error) {
	return provider.Result{Status: f.status, Detail: f.detail}, nil
}

func testEngine() *engine.Engine {
	return engine.New(
		fakeProvider{id: "a", status: provider.StatusTaken, detail: "found"},
		fakeProvider{id: "b", status: provider.StatusAvailable, detail: "free"},
	)
}

// TestRunCheckEmitsJSONArray pins the one-shot shape: stdout is a single JSON
// array of every buffered result, not newline-delimited objects.
func TestRunCheckEmitsJSONArray(t *testing.T) {
	var buf bytes.Buffer
	if err := runCheck(context.Background(), &buf, testEngine(), "acema", ""); err != nil {
		t.Fatalf("runCheck: %v", err)
	}

	out := buf.String()
	if !strings.HasPrefix(strings.TrimSpace(out), "[") {
		t.Fatalf("output is not a JSON array: %q", out)
	}
	if !strings.HasSuffix(out, "]\n") {
		t.Fatalf("output does not end with a newline-terminated array: %q", out)
	}

	var got []engine.Result
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not one JSON value: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d results, want 2: %+v", len(got), got)
	}
	for i, want := range []string{"a", "b"} {
		if got[i].ID != want {
			t.Errorf("result[%d].ID = %q, want %q", i, got[i].ID, want)
		}
	}
}

// TestRunCheckSQLiteWritesFile checks the --out path buffers the scan into the
// SQLite report and leaves stdout untouched.
func TestRunCheckSQLiteWritesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "checks.sqlite")
	var buf bytes.Buffer
	if err := runCheck(context.Background(), &buf, testEngine(), "acema", path); err != nil {
		t.Fatalf("runCheck: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("--out wrote %q to stdout, want nothing", buf.String())
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open report: %v", err)
	}
	defer db.Close()

	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM checks").Scan(&n); err != nil {
		t.Fatalf("count checks: %v", err)
	}
	if n != 2 {
		t.Fatalf("checks rows = %d, want 2", n)
	}
}

func TestParseCheck(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantLabel string
		wantOut   string
		wantErr   bool
	}{
		{name: "label only", args: []string{"acema"}, wantLabel: "acema"},
		{name: "with out", args: []string{"acema", "--out", "report.sqlite"}, wantLabel: "acema", wantOut: "report.sqlite"},
		{name: "empty", args: nil, wantErr: true},
		{name: "dangling out", args: []string{"acema", "--out"}, wantErr: true},
		{name: "unknown flag", args: []string{"acema", "--nope"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label, out, err := parseCheck(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseCheck(%v) = %q, %q; want error", tt.args, label, out)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseCheck(%v): %v", tt.args, err)
			}
			if label != tt.wantLabel || out != tt.wantOut {
				t.Errorf("parseCheck(%v) = %q, %q; want %q, %q", tt.args, label, out, tt.wantLabel, tt.wantOut)
			}
		})
	}
}

// blockingProvider never resolves on its own; it waits for the scan context to
// be canceled, mirroring a hung source.
type blockingProvider struct{ fakeProvider }

func (b blockingProvider) Check(ctx context.Context, _ string) (provider.Result, error) {
	<-ctx.Done()
	return provider.Result{}, ctx.Err()
}

// TestRunCheckHonorsScanDeadline pins that a one-shot check stops at the scan
// deadline instead of hanging on a source that never answers.
func TestRunCheckHonorsScanDeadline(t *testing.T) {
	eng := engine.New(blockingProvider{fakeProvider{id: "slow"}})
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	var buf bytes.Buffer
	done := make(chan error, 1)
	go func() { done <- runCheck(ctx, &buf, eng, "acema", "") }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("runCheck: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("runCheck did not return after the scan deadline")
	}

	var got []engine.Result
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not JSON: %v (%q)", err, buf.String())
	}
	if len(got) != 0 {
		t.Errorf("got %+v, want no results (the blocked provider was abandoned)", got)
	}
}

// TestWithScanTimeoutDeadline pins that the serve middleware hands the handler
// a context that reaches its deadline.
func TestWithScanTimeoutDeadline(t *testing.T) {
	var deadline bool
	h := withScanTimeout(50*time.Millisecond, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		deadline = r.Context().Err() == context.DeadlineExceeded
		_, _ = w.Write([]byte("late"))
	}))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/check", nil))
	if !deadline {
		t.Fatal("handler context did not reach its deadline")
	}
}
