package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/petermichon/domain-browser/server/internal/fixtures"
)

// exportLabelPattern mirrors api's label rule. Labels are also directory names,
// so reject anything that is not a safe DNS label.
var exportLabelPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)

// labelList collects repeated --label flags in order.
type labelList []string

func (l *labelList) String() string { return strings.Join(*l, ",") }

func (l *labelList) Set(value string) error {
	*l = append(*l, value)
	return nil
}

// exportFixtures writes the source inventory and, for each --label, one raw
// HTTP fixture per source. With no labels it writes only sources.json.
func exportFixtures(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	out := fs.String("out", "fixtures", "directory for sources.json and label captures")
	concurrency := fs.Int("concurrency", 8, "maximum simultaneous source checks")
	var labels labelList
	fs.Var(&labels, "label", "label to capture (repeatable)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("export: unexpected argument %q", fs.Arg(0))
	}
	if strings.TrimSpace(*out) == "" {
		return fmt.Errorf("export: --out must not be empty")
	}
	if *concurrency < 1 {
		return fmt.Errorf("export: --concurrency must be positive")
	}

	normalized := make([]string, 0, len(labels))
	seen := make(map[string]bool, len(labels))
	for _, raw := range labels {
		label := strings.ToLower(strings.TrimSpace(raw))
		if !exportLabelPattern.MatchString(label) {
			return fmt.Errorf("export: invalid label %q", raw)
		}
		if seen[label] {
			continue
		}
		seen[label] = true
		normalized = append(normalized, label)
	}

	summary, err := fixtures.Export(context.Background(), fixtures.ExportOptions{
		OutDir:      *out,
		Labels:      normalized,
		Concurrency: *concurrency,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "sources: %d (%d enabled, %d keyed) -> %s\n",
		summary.SourceCount, summary.EnabledCount, summary.KeyedCount, summary.SourcesPath)
	for _, ls := range summary.Labels {
		fmt.Fprintf(stdout, "fixtures %q: %d captured, %d errors -> %s\n",
			ls.Label, ls.Fixtures, ls.Errors, ls.Dir)
	}
	return nil
}
