package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

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
	if err := runCheck(&buf, testEngine(), "acema", ""); err != nil {
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
	if err := runCheck(&buf, testEngine(), "acema", path); err != nil {
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
