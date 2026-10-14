package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/petermichon/domain-browser/backend/internal/engine"
	"github.com/petermichon/domain-browser/backend/internal/provider"
	"github.com/petermichon/domain-browser/backend/internal/sink"
	"github.com/petermichon/domain-browser/backend/internal/store"
)

const usage = `usage: domainbrowser <command> [args]

Commands:
  serve                              start the HTTP server
  check <label> [--out <file.sqlite>] check a single label across the sources
                                     (--out writes a SQLite report instead of JSON)
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
	return fmt.Errorf("serve: not implemented")
}

func check(args []string) error {
	label, out, err := parseCheck(args)
	if err != nil {
		return err
	}

	eng := engine.NewWithStore(store.NewMemory(), provider.Default().All()...)

	if out != "" {
		return checkSQLite(eng, label, out)
	}

	enc := json.NewEncoder(os.Stdout)
	for res := range eng.Scan(context.Background(), label) {
		if err := enc.Encode(res); err != nil {
			return err
		}
	}
	return nil
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

// checkSQLite runs the scan and upserts every result into the SQLite report.
func checkSQLite(eng *engine.Engine, label, path string) error {
	s, err := sink.OpenSQLite(path)
	if err != nil {
		return err
	}

	results := make([]engine.Result, 0, len(eng.Providers()))
	for res := range eng.Scan(context.Background(), label) {
		results = append(results, res)
	}
	if err := s.Write(context.Background(), results); err != nil {
		s.Close()
		return err
	}
	return s.Close()
}
