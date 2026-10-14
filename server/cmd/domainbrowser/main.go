package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
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
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	switch args[0] {
	case "serve":
		if err := serve(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "domainbrowser: %v\n", err)
			os.Exit(1)
		}
	case "check":
		if err := check(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "domainbrowser: %v\n", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "domainbrowser: unknown command %q\n\n%s", args[0], usage)
		os.Exit(2)
	}
}

func serve(args []string) error {
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

	// A long-lived engine over one memory store so cached verdicts survive
	// across requests (the point of serve, versus the one-shot check).
	eng := engine.NewWithStore(store.NewMemory(), provider.Default().All()...)
	srv := &http.Server{
		Addr:    *addr,
		Handler: api.New(eng, *origin).Handler(),
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()
	fmt.Fprintf(os.Stderr, "domainbrowser: listening on %s (CORS origin %s)\n", *addr, *origin)

	select {
	case err := <-errc:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

func check(args []string) error {
	label, out, err := parseCheck(args)
	if err != nil {
		return err
	}

	eng := engine.NewWithStore(store.NewMemory(), provider.Default().All()...)
	return runCheck(os.Stdout, eng, label, out)
}

// runCheck performs a one-shot scan and writes the whole buffered result set:
// a single JSON array for stdout, or a SQLite report when out is set. Unlike
// serve, no result is emitted until every provider has resolved.
func runCheck(w io.Writer, eng *engine.Engine, label, out string) error {
	results := eng.ScanAll(context.Background(), label)
	if out != "" {
		return checkSQLite(results, out)
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
func checkSQLite(results []engine.Result, path string) error {
	s, err := sink.OpenSQLite(path)
	if err != nil {
		return err
	}

	if err := s.Write(context.Background(), results); err != nil {
		s.Close()
		return err
	}
	return s.Close()
}
