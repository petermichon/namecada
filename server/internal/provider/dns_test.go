package provider

import (
	"context"
	"errors"
	"net"
	"testing"
)

// resolverFunc adapts a function to a Resolver so a test can both answer a
// lookup and capture the host it was asked for.
type resolverFunc func(context.Context, string) ([]string, error)

func (f resolverFunc) LookupHost(ctx context.Context, host string) ([]string, error) {
	return f(ctx, host)
}

// TestDNSCheck pins the DNS verdict mapping: a not-found DNS error is
// available, a successful lookup is taken, and every other failure (timeout,
// temporary, a non-DNS error) is unknown.
func TestDNSCheck(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want Status
	}{
		{"taken", nil, StatusTaken},
		{"available", &net.DNSError{Err: "no such host", IsNotFound: true}, StatusAvailable},
		{"timeout", &net.DNSError{Err: "i/o timeout", IsTimeout: true}, StatusUnknown},
		{"temporary", &net.DNSError{Err: "server misbehaving", IsTemporary: true}, StatusUnknown},
		{"non-dns error", errors.New("boom"), StatusUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newDNS("test", "test", "example.test")
			p.resolver = resolverFunc(func(context.Context, string) ([]string, error) {
				if tt.err != nil {
					return nil, tt.err
				}
				return []string{"192.0.2.1"}, nil
			})

			got, err := p.Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
		})
	}
}

// TestDNSLookupHost pins the exact host the provider resolves and the public
// URL it exposes.
func TestDNSLookupHost(t *testing.T) {
	var gotHost string
	p := newDNS("euorg", "euorg", "eu.org")
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
	if gotHost != "acema.eu.org" {
		t.Errorf("looked up %q, want %q", gotHost, "acema.eu.org")
	}
	if got := p.URL("acema"); got != "https://acema.eu.org/" {
		t.Errorf("URL = %q, want %q", got, "https://acema.eu.org/")
	}
	if p.ID() != "euorg" || p.Platform() != "euorg" {
		t.Errorf("metadata = %q/%q, want euorg/euorg", p.ID(), p.Platform())
	}
	if got := p.Name("acema"); got != "acema" {
		t.Errorf("Name = %q, want %q", got, "acema")
	}
}

// TestDNSDefaultResolver pins that a provider with no injected resolver falls
// back to net.DefaultResolver.
func TestDNSDefaultResolver(t *testing.T) {
	p := newDNS("test", "test", "example.test")
	if _, ok := p.resolverOrDefault().(*net.Resolver); !ok {
		t.Errorf("resolverOrDefault = %T, want *net.Resolver", p.resolverOrDefault())
	}
}
