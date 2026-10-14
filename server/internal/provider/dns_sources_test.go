package provider

import (
	"context"
	"testing"
)

// TestDNSSources pins each free-subdomain provider's host, public URL, metadata
// and display name. Every one resolves `{label}.{host}` and is free.
func TestDNSSources(t *testing.T) {
	cases := []struct {
		name     string
		build    func() Provider
		id       string
		platform string
		host     string
		display  string
	}{
		{"euorg", NewEUOrg, "euorg", "euorg", "eu.org", "eu.org"},
		{"duckdns", NewDuckDNS, "duckdns", "duckdns", "duckdns.org", "DuckDNS"},
		{"noip", NewNoIP, "noip", "noip", "no-ip.org", "No-IP"},
		{"dynv6", NewDynV6, "dynv6", "dynv6", "dynv6.net", "dynv6"},
		{"notedlol", NewNotedLol, "notedlol", "notedlol", "noted.lol", "noted.lol"},
		{"isagooddev", NewIsAGoodDev, "isagooddev", "isagooddev", "is-a-good.dev", "is-a-good.dev"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := tc.build().(*dns)
			var gotHost string
			p.resolver = resolverFunc(func(_ context.Context, host string) ([]string, error) {
				gotHost = host
				return []string{"192.0.2.1"}, nil
			})

			got, err := p.Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != StatusTaken {
				t.Errorf("status = %q, want %q", got.Status, StatusTaken)
			}
			if gotHost != "acema."+tc.host {
				t.Errorf("looked up %q, want %q", gotHost, "acema."+tc.host)
			}
			if want := "https://acema." + tc.host + "/"; p.URL("acema") != want {
				t.Errorf("URL = %q, want %q", p.URL("acema"), want)
			}
			if p.ID() != tc.id || p.Platform() != tc.platform {
				t.Errorf("metadata = %q/%q, want %q/%q", p.ID(), p.Platform(), tc.id, tc.platform)
			}
			if got := p.Name("acema"); got != "acema" {
				t.Errorf("Name = %q, want %q", got, "acema")
			}
			if got := PlatformName(tc.platform); got != tc.display {
				t.Errorf("PlatformName(%q) = %q, want %q", tc.platform, got, tc.display)
			}
			if Paid(p) {
				t.Errorf("Paid = true, want free for %q", tc.id)
			}
		})
	}
}
