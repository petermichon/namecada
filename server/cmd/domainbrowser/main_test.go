package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
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
		{name: "empty out", args: []string{"acema", "--out", ""}, wantErr: true},
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

// withTestEngine swaps the package-level engine factory for a deterministic one
// so the command wiring never touches the network.
func withTestEngine(t *testing.T, eng *engine.Engine) {
	t.Helper()
	old := newEngine
	newEngine = func(provider.Config) *engine.Engine { return eng }
	t.Cleanup(func() { newEngine = old })
}

func testResults() []engine.Result {
	return []engine.Result{
		{ID: "a", Platform: "p-a", Name: "a:acema", URL: "https://a.test/acema", Status: provider.StatusTaken},
	}
}

// TestRunDispatchesCommands pins the CLI wiring: every command maps to the
// expected exit code and writes to the stream it is handed.
func TestRunDispatchesCommands(t *testing.T) {
	withTestEngine(t, testEngine())

	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{name: "no args", wantCode: 2, wantStderr: "usage:"},
		{name: "help", args: []string{"help"}, wantCode: 0, wantStdout: "usage:"},
		{name: "unknown command", args: []string{"frobnicate"}, wantCode: 2, wantStderr: "unknown command"},
		{name: "check", args: []string{"check", "acema"}, wantCode: 0, wantStdout: "["},
		{name: "check without label", args: []string{"check"}, wantCode: 1, wantStderr: "requires a <label>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if got := run(tt.args, &stdout, &stderr); got != tt.wantCode {
				t.Fatalf("run(%v) = %d, want %d", tt.args, got, tt.wantCode)
			}
			if !strings.Contains(stdout.String(), tt.wantStdout) {
				t.Errorf("stdout = %q, want to contain %q", stdout.String(), tt.wantStdout)
			}
			if !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Errorf("stderr = %q, want to contain %q", stderr.String(), tt.wantStderr)
			}
		})
	}
}

// TestServeFlagParsing pins that serve rejects a bad flag or a stray argument
// before it ever binds a listener.
func TestServeFlagParsing(t *testing.T) {
	withTestEngine(t, testEngine())

	for _, tt := range []struct {
		name string
		args []string
	}{
		{name: "unknown flag", args: []string{"--nope"}},
		{name: "unexpected argument", args: []string{"extra"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := serve(provider.Config{}, tt.args); err == nil {
				t.Fatalf("serve(%v) = nil, want error", tt.args)
			}
		})
	}
}

// TestServeContextListenError pins that a listener failure is surfaced.
func TestServeContextListenError(t *testing.T) {
	withTestEngine(t, testEngine())
	if err := serveContext(context.Background(), provider.Config{}, "127.0.0.1:99999", ""); err == nil {
		t.Fatal("serveContext = nil, want a listen error")
	}
}

// TestServeContextShutsDownOnCancel pins that the serve loop returns cleanly
// once its context is canceled.
func TestServeContextShutsDownOnCancel(t *testing.T) {
	withTestEngine(t, testEngine())

	ctx, cancel := context.WithCancel(context.Background())
	errc := make(chan error, 1)
	go func() { errc <- serveContext(ctx, provider.Config{}, "127.0.0.1:0", "") }()

	// Let the listener bind before asking it to stop.
	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-errc:
		if err != nil {
			t.Fatalf("serveContext: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("serveContext did not return after cancel")
	}
}

// TestServeWithLoggerLogsLifecycle pins that the serve loop logs a startup line
// and, on shutdown, a shutting-down and a stopped line.
func TestServeWithLoggerLogsLifecycle(t *testing.T) {
	withTestEngine(t, testEngine())

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	ctx, cancel := context.WithCancel(context.Background())
	errc := make(chan error, 1)
	go func() { errc <- serveWithLogger(ctx, provider.Config{}, "127.0.0.1:0", "", logger) }()

	// Let the listener bind before asking it to stop.
	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-errc:
		if err != nil {
			t.Fatalf("serveWithLogger: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("serveWithLogger did not return after cancel")
	}

	logged := buf.String()
	for _, want := range []string{"msg=listening", `msg="shutting down"`, "msg=stopped"} {
		if !strings.Contains(logged, want) {
			t.Errorf("log = %q, want to contain %q", logged, want)
		}
	}
}

// TestCheckSQLiteOpenError pins that an unwritable report path is reported.
func TestCheckSQLiteOpenError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "report.sqlite")
	if err := checkSQLite(context.Background(), testResults(), path); err == nil {
		t.Fatal("checkSQLite = nil, want an open error")
	}
}

// TestCheckSQLiteWriteError pins that a canceled context aborts the write.
func TestCheckSQLiteWriteError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.sqlite")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := checkSQLite(ctx, testResults(), path); err == nil {
		t.Fatal("checkSQLite = nil, want a write error")
	}
}

// TestConfigFromEnv pins the DOMAINBROWSER_<SOURCE>_<KIND> mapping.
func TestConfigFromEnv(t *testing.T) {
	t.Setenv("DOMAINBROWSER_DISCOGS_TOKEN", "d")
	t.Setenv("DOMAINBROWSER_X_BEARER_TOKEN", "x")
	t.Setenv("DOMAINBROWSER_TWITCH_CLIENT_ID", "ti")
	t.Setenv("DOMAINBROWSER_TWITCH_CLIENT_SECRET", "ts")
	t.Setenv("DOMAINBROWSER_COMPANIESHOUSE_KEY", "c")
	t.Setenv("DOMAINBROWSER_OPENCORPORATES_TOKEN", "o")

	want := provider.Config{
		DiscogsToken:        "d",
		XBearerToken:        "x",
		TwitchClientID:      "ti",
		TwitchClientSecret:  "ts",
		CompaniesHouseKey:   "c",
		OpenCorporatesToken: "o",
	}
	if got := configFromEnv(); got != want {
		t.Errorf("configFromEnv() = %+v, want %+v", got, want)
	}
}

// TestCheckOutputHasNoCredentials pins that a configured credential never
// reaches the emitted JSON.
func TestCheckOutputHasNoCredentials(t *testing.T) {
	const secret = "leaky-credential-token"
	t.Setenv("DOMAINBROWSER_DISCOGS_TOKEN", secret)
	withTestEngine(t, testEngine())

	var stdout, stderr bytes.Buffer
	if code := run([]string{"check", "acema"}, &stdout, &stderr); code != 0 {
		t.Fatalf("run = %d, want 0; stderr=%q", code, stderr.String())
	}
	if strings.Contains(stdout.String(), secret) || strings.Contains(stderr.String(), secret) {
		t.Errorf("credential leaked into output: stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

// TestServeLogsNoCredentials pins that startup logs source names, never values.
func TestServeLogsNoCredentials(t *testing.T) {
	const secret = "leaky-credential-token"
	t.Setenv("DOMAINBROWSER_DISCOGS_TOKEN", secret)
	withTestEngine(t, testEngine())

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	ctx, cancel := context.WithCancel(context.Background())
	errc := make(chan error, 1)
	go func() { errc <- serveWithLogger(ctx, configFromEnv(), "127.0.0.1:0", "", logger) }()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-errc:
		if err != nil {
			t.Fatalf("serveWithLogger: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("serveWithLogger did not return after cancel")
	}

	if strings.Contains(buf.String(), secret) {
		t.Errorf("credential leaked into logs: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "msg=sources") {
		t.Errorf("log = %q, want a sources line", buf.String())
	}
}
