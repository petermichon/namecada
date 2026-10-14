package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestParseBootstrap(t *testing.T) {
	t.Run("maps TLDs to their registry base URL, normalizing the trailing slash", func(t *testing.T) {
		map0 := parseBootstrap([]byte(`{"services":[
			[["com","net"],["https://rdap.verisign.com/com/v1/"]],
			[["test"],["https://rdap.test"]]
		]}`))
		if got := map0["com"]; got != "https://rdap.verisign.com/com/v1/" {
			t.Errorf("com = %q", got)
		}
		if got := map0["net"]; got != "https://rdap.verisign.com/com/v1/" {
			t.Errorf("net = %q", got)
		}
		if got := map0["test"]; got != "https://rdap.test/" {
			t.Errorf("test = %q", got)
		}
	})

	t.Run("ignores malformed services", func(t *testing.T) {
		got := parseBootstrap([]byte(`{"services":[["io",["https://rdap.example/"]],[["no-urls"],[]],"nope",null]}`))
		if len(got) != 0 {
			t.Errorf("map = %v, want empty", got)
		}
	})

	t.Run("returns an empty map for non-object input", func(t *testing.T) {
		if len(parseBootstrap([]byte(`null`))) != 0 {
			t.Error("null did not yield an empty map")
		}
		if len(parseBootstrap([]byte(`{}`))) != 0 {
			t.Error("{} did not yield an empty map")
		}
		if len(parseBootstrap([]byte(`not json`))) != 0 {
			t.Error("invalid JSON did not yield an empty map")
		}
	})
}

func TestBootstrapResolve(t *testing.T) {
	t.Run("fetches once and caches for the TTL", func(t *testing.T) {
		var calls int
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			_, _ = w.Write([]byte(`{"services":[[["com"],["https://custom.example/rdap/"]]]}`))
		}))
		defer srv.Close()

		now := time.Unix(1700000000, 0)
		bs := &Bootstrap{
			URL:    srv.URL,
			Client: srv.Client(),
			TTL:    time.Hour,
			Now:    func() time.Time { return now },
		}

		got, ok := bs.Resolve(context.Background(), ".com")
		if !ok || got != "https://custom.example/rdap/" {
			t.Fatalf("Resolve(.com) = %q, %v", got, ok)
		}
		if _, ok := bs.Resolve(context.Background(), "com"); !ok {
			t.Fatal("Resolve(com) did not find the cached entry")
		}
		if calls != 1 {
			t.Fatalf("fetches = %d, want 1", calls)
		}

		now = now.Add(2 * time.Hour)
		if _, ok := bs.Resolve(context.Background(), "com"); !ok {
			t.Fatal("Resolve after expiry failed")
		}
		if calls != 2 {
			t.Fatalf("fetches after expiry = %d, want 2", calls)
		}
	})

	t.Run("falls back when the bootstrap is unreachable", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		url := srv.URL
		srv.Close() // no listener: connection refused

		bs := &Bootstrap{URL: url, Client: http.DefaultClient}
		if got, ok := bs.Resolve(context.Background(), "com"); !ok || got != fallbackBases["com"] {
			t.Errorf("Resolve(com) = %q, %v, want fallback", got, ok)
		}
		if _, ok := bs.Resolve(context.Background(), "org"); ok {
			t.Error("Resolve(org) unexpectedly resolved with only the fallback map")
		}
	})

	t.Run("layers the bootstrap map over the fallback registries", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"services":[[["com"],["https://override.example/"]]]}`))
		}))
		defer srv.Close()

		bs := &Bootstrap{URL: srv.URL, Client: srv.Client()}
		got, ok := bs.Resolve(context.Background(), "com")
		if !ok || got != "https://override.example/" {
			t.Fatalf("Resolve(com) = %q, %v, want the bootstrap override", got, ok)
		}
		if got, ok := bs.Resolve(context.Background(), "io"); !ok || got != fallbackBases["io"] {
			t.Fatalf("Resolve(io) = %q, %v, want the fallback", got, ok)
		}
	})
}

func TestDomainCheck(t *testing.T) {
	tests := []struct {
		name string
		code int
		want Status
	}{
		{"registered", http.StatusOK, StatusTaken},
		{"unregistered", http.StatusNotFound, StatusAvailable},
		{"server error", http.StatusInternalServerError, StatusUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				w.WriteHeader(tt.code)
				_, _ = w.Write([]byte(`{"events":[]}`))
			}))
			defer srv.Close()

			d := &Domain{TLD: ".dev", BaseURL: srv.URL + "/", Client: srv.Client()}
			got, err := d.Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if gotPath != "/domain/acema.dev" {
				t.Errorf("path = %q, want %q", gotPath, "/domain/acema.dev")
			}
		})
	}
}

func TestDomainCheckReservedLabels(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		calls++
	}))
	defer srv.Close()

	d := &Domain{TLD: ".com", BaseURL: srv.URL + "/", Client: srv.Client()}
	for _, label := range []string{"example", "test", "invalid", "localhost", "EXAMPLE"} {
		got, err := d.Check(context.Background(), label)
		if err != nil {
			t.Fatalf("Check(%q) returned error: %v", label, err)
		}
		if got.Status != StatusTaken {
			t.Errorf("Check(%q) status = %q, want %q", label, got.Status, StatusTaken)
		}
	}
	if calls != 0 {
		t.Errorf("reserved labels made %d requests, want 0", calls)
	}
}

func TestDomainCheckUnsupportedTLD(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close() // force the bootstrap fetch to fail

	d := &Domain{TLD: ".zzz", Bootstrap: &Bootstrap{URL: url, Client: http.DefaultClient}}
	got, err := d.Check(context.Background(), "acema")
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if got.Status != StatusUnknown {
		t.Errorf("status = %q, want %q", got.Status, StatusUnknown)
	}
	if got.Detail != "no RDAP endpoint" {
		t.Errorf("detail = %q, want %q", got.Detail, "no RDAP endpoint")
	}
}

func TestDomainCheckResolvesViaBootstrap(t *testing.T) {
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/domain/acema.com" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/domain/acema.com")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer registry.Close()

	bootstrap := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"services":[[["com"],["` + registry.URL + `/"]]]}`))
	}))
	defer bootstrap.Close()

	d := &Domain{
		TLD:       ".com",
		Bootstrap: &Bootstrap{URL: bootstrap.URL, Client: bootstrap.Client()},
		Client:    registry.Client(),
	}
	got, err := d.Check(context.Background(), "acema")
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if got.Status != StatusTaken {
		t.Errorf("status = %q, want %q", got.Status, StatusTaken)
	}
}

func TestBootstrapUserAgent(t *testing.T) {
	var gotUA string
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer registry.Close()

	bootstrap := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = w.Write([]byte(`{"services":[[["com"],["` + registry.URL + `/"]]]}`))
	}))
	defer bootstrap.Close()

	d := &Domain{
		TLD:       ".com",
		Bootstrap: &Bootstrap{URL: bootstrap.URL, Client: bootstrap.Client()},
		Client:    registry.Client(),
	}
	if _, err := d.Check(context.Background(), "acema"); err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if gotUA != UserAgent {
		t.Errorf("bootstrap User-Agent = %q, want %q", gotUA, UserAgent)
	}
}

func TestDomainMinLength(t *testing.T) {
	if got := NewDomain(".dev").MinLength(); got != 2 {
		t.Errorf(".dev MinLength = %d, want 2", got)
	}
	for _, tld := range []string{".com", ".org", ".io", ".ai", ".app"} {
		if got := NewDomain(tld).MinLength(); got != 0 {
			t.Errorf("%s MinLength = %d, want 0", tld, got)
		}
	}
}

func TestDomainsMetadata(t *testing.T) {
	tests := []struct {
		tld  string
		id   string
		name string
		url  string
	}{
		{".com", "domain.com", "acema.com", "https://acema.com"},
		{".dev", "domain.dev", "acema.dev", "https://acema.dev"},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			d := NewDomain(tt.tld)
			if d.ID() != tt.id {
				t.Errorf("ID = %q, want %q", d.ID(), tt.id)
			}
			if d.Platform() != "domains" {
				t.Errorf("Platform = %q, want %q", d.Platform(), "domains")
			}
			if got := d.Name("acema"); got != tt.name {
				t.Errorf("Name = %q, want %q", got, tt.name)
			}
			if got := d.URL("acema"); got != tt.url {
				t.Errorf("URL = %q, want %q", got, tt.url)
			}
		})
	}
}

func TestNewDomains(t *testing.T) {
	got := NewDomains()
	want := []string{
		"domain.com", "domain.org", "domain.dev",
		"domain.io", "domain.ai", "domain.app",
	}
	if len(got) != len(want) {
		t.Fatalf("NewDomains() returned %d providers, want %d", len(got), len(want))
	}
	for i, p := range got {
		if p.ID() != want[i] {
			t.Errorf("provider %d ID = %q, want %q", i, p.ID(), want[i])
		}
	}
}
