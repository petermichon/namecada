package provider

import (
	"context"
	"errors"
	"net"
)

// Resolver resolves a host name to addresses. *net.Resolver satisfies it, and
// tests substitute a fake so a verdict can be pinned without touching DNS.
type Resolver interface {
	// LookupHost returns the addresses the host resolves to. A CNAME chain is
	// followed, so an A, AAAA or CNAME record all yield a non-empty result.
	LookupHost(ctx context.Context, host string) ([]string, error)
}

// dns is a provider whose presence signal is DNS: it resolves
// `{label}.{host}` for a free-subdomain namespace (eu.org, duckdns.org, ...)
// that does not serve wildcard DNS, so only claimed names resolve. A
// not-found error means the name is free, a successful lookup means it is
// taken, and any other failure (timeout, temporary, a non-DNS error) is
// unknown, which is never cached and so self-heals.
type dns struct {
	base
	// host is the parent domain the label is prepended to, without a dot.
	host string
	// resolver answers the lookup; nil falls back to net.DefaultResolver.
	resolver Resolver
}

// newDNS builds a DNS provider for host. The public page is
// https://{label}.{host}/, matching the HTTP host-status family.
func newDNS(id, platform, host string) *dns {
	return &dns{
		base: base{
			id:       id,
			platform: platform,
			urlFmt:   "https://%s." + host + "/",
		},
		host: host,
	}
}

// resolverOrDefault returns the injected resolver, or net.DefaultResolver when
// none was supplied.
func (d *dns) resolverOrDefault() Resolver {
	if d.resolver != nil {
		return d.resolver
	}
	return net.DefaultResolver
}

// Check resolves {label}.{host}. A DNSError with IsNotFound maps to available,
// a successful lookup to taken, and every other error to unknown. It never
// returns an error: a transient failure must not be cached as a verdict.
func (d *dns) Check(ctx context.Context, label string) (Result, error) {
	_, err := d.resolverOrDefault().LookupHost(ctx, label+"."+d.host)
	if err == nil {
		return Result{Status: StatusTaken}, nil
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
		return Result{Status: StatusAvailable}, nil
	}
	return Result{Status: StatusUnknown}, nil
}
