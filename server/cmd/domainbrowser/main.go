package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/petermichon/domain-browser/server/internal/api"
	"github.com/petermichon/domain-browser/server/internal/engine"
	"github.com/petermichon/domain-browser/server/internal/provider"
	"github.com/petermichon/domain-browser/server/internal/sink"
	"github.com/petermichon/domain-browser/server/internal/store"
)

const usage = `usage: domainbrowser <command> [args]

Commands:
  serve [--addr :8080] [--origin URL] start the HTTP server (GET /check?name=)
  check <label> [--out <file.sqlite>]  check a single label across the sources,
                                      printing one JSON array of results
                                      (--out writes a SQLite report instead)
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run dispatches a command line, returning the process exit code. It is split
// out of main so the CLI wiring can be exercised without exiting the test
// process.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}

	// Read keyed-source credentials once, at the edge; the provider package
	// never touches the environment.
	cfg := configFromEnv()

	switch args[0] {
	case "serve":
		if err := serve(cfg, args[1:]); err != nil {
			fmt.Fprintf(stderr, "domainbrowser: %v\n", err)
			return 1
		}
	case "check":
		if err := check(cfg, args[1:], stdout); err != nil {
			fmt.Fprintf(stderr, "domainbrowser: %v\n", err)
			return 1
		}
	case "help", "-h", "--help":
		fmt.Fprint(stdout, usage)
	default:
		fmt.Fprintf(stderr, "domainbrowser: unknown command %q\n\n%s", args[0], usage)
		return 2
	}
	return 0
}

// configFromEnv reads keyed-source credentials from the
// DOMAINBROWSER_<SOURCE>_<KIND> environment variables. An unset variable leaves
// its field empty, which omits the corresponding source from the registry.
func configFromEnv() provider.Config {
	return provider.Config{
		DiscogsToken:        os.Getenv("DOMAINBROWSER_DISCOGS_TOKEN"),
		TwitchClientID:      os.Getenv("DOMAINBROWSER_TWITCH_CLIENT_ID"),
		TwitchClientSecret:  os.Getenv("DOMAINBROWSER_TWITCH_CLIENT_SECRET"),
		CompaniesHouseKey:   os.Getenv("DOMAINBROWSER_COMPANIESHOUSE_KEY"),
		OpenCorporatesToken: os.Getenv("DOMAINBROWSER_OPENCORPORATES_TOKEN"),
	}
}

// newEngine builds the engine both commands scan with. It is a variable so
// tests can substitute a deterministic engine that never touches the network.
var newEngine = func(cfg provider.Config) *engine.Engine {
	return engine.NewWithStore(store.NewMemory(), provider.Default(cfg).All()...)
}

func serve(cfg provider.Config, args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	addr := fs.String("addr", ":8080", "address to listen on")
	origin := fs.String("origin", api.DefaultOrigin, "CORS origin allowed to call the API")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("serve: unexpected argument %q", fs.Arg(0))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return serveContext(ctx, cfg, *addr, *origin)
}

// serveContext runs the HTTP server until ctx is canceled or the listener
// fails, then shuts it down. It is separated from serve so tests can drive the
// shutdown path with their own context.
func serveContext(ctx context.Context, cfg provider.Config, addr, origin string) error {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	return serveWithLogger(ctx, cfg, addr, origin, logger)
}

// serveWithLogger runs the HTTP server with logger, emitting a structured
// request log per call plus startup and shutdown lines. It is separated from
// serveContext so tests can capture its logs.
func serveWithLogger(ctx context.Context, cfg provider.Config, addr, origin string, logger *slog.Logger) error {
	// A long-lived engine over one memory store so cached verdicts survive
	// across requests (the point of serve, versus the one-shot check).
	eng := newEngine(cfg)
	srv := &http.Server{
		Addr:    addr,
		Handler: withScanTimeout(engine.ScanTimeout, api.New(eng, origin, logger).Handler()),
	}

	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()
	// Log which optional (keyed) sources are live by name only — never their
	// credential values.
	logger.Info("sources",
		"keyed", provider.KeyedNames(cfg),
		"total", len(eng.Providers()),
	)
	logger.Info("listening", "addr", addr, "origin", origin)

	select {
	case err := <-errc:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		logger.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			logger.Error("shutdown failed", "error", err)
			return err
		}
		logger.Info("stopped")
		return nil
	}
}

func check(cfg provider.Config, args []string, w io.Writer) error {
	label, out, err := parseCheck(args)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), engine.ScanTimeout)
	defer cancel()

	return runCheck(ctx, w, newEngine(cfg), label, out)
}

// withScanTimeout bounds every served scan with a context deadline, so a hung
// source cannot hold a request open past the timeout. The deadline flows to the
// engine, which stops emitting once it fires.
func withScanTimeout(timeout time.Duration, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// runCheck performs a one-shot scan and writes the whole buffered result set:
// a single JSON array for stdout, or a SQLite report when out is set. Unlike
// serve, no result is emitted until every provider has resolved.
func runCheck(ctx context.Context, w io.Writer, eng *engine.Engine, label, out string) error {
	results := eng.ScanAll(ctx, label)
	if out != "" {
		return checkSQLite(ctx, results, out)
	}
	return writeJSON(w, results)
}

// writeJSON encodes results as a single JSON array (plus a trailing newline).
func writeJSON(w io.Writer, results []engine.Result) error {
	return json.NewEncoder(w).Encode(results)
}

// parseCheck splits "check" arguments into the label and an optional --out path.
func parseCheck(args []string) (label, out string, err error) {
	if len(args) == 0 || args[0] == "" {
		return "", "", fmt.Errorf("check requires a <label> argument")
	}
	label = args[0]

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--out":
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("check: --out requires a <file.sqlite> argument")
			}
			out = args[i+1]
			i++
		default:
			return "", "", fmt.Errorf("check: unexpected argument %q", args[i])
		}
	}
	if out == "" && len(args) > 1 {
		return "", "", fmt.Errorf("check: --out requires a <file.sqlite> argument")
	}
	return label, out, nil
}

// checkSQLite upserts every buffered result into the SQLite report.
func checkSQLite(ctx context.Context, results []engine.Result, path string) error {
	s, err := sink.OpenSQLite(path)
	if err != nil {
		return err
	}

	if err := s.Write(ctx, results); err != nil {
		s.Close()
		return err
	}
	return s.Close()
}
