package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/petermichon/domain-browser/backend/internal/engine"
	"github.com/petermichon/domain-browser/backend/internal/provider"
	"github.com/petermichon/domain-browser/backend/internal/store"
)

const usage = `usage: domainbrowser <command> [args]

Commands:
  serve           start the HTTP server
  check <label>   check a single label across the sources
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
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "domainbrowser: check requires a <label> argument")
			os.Exit(2)
		}
		if err := check(args[1]); err != nil {
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

func check(label string) error {
	eng := engine.NewWithStore(store.NewMemory(), provider.Default().All()...)

	enc := json.NewEncoder(os.Stdout)
	for res := range eng.Scan(context.Background(), label) {
		if err := enc.Encode(res); err != nil {
			return err
		}
	}
	return nil
}
