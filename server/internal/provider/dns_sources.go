package provider

// The free-subdomain group below checks a name by resolving `{label}.{host}`.
// None of these parents serve wildcard DNS, so only a claimed name resolves: a
// successful lookup means taken and an NXDOMAIN means free. The label is sent
// to the DNS resolver, so a scan leaks the name being checked (the same caveat
// as any DNS lookup). Each provider reuses the shared `dns` helper (dns.go).

// NewEUOrg checks a free subdomain under eu.org. eu.org has no wildcard DNS, so
// only claimed names resolve: `{label}.eu.org` resolving means taken and an
// NXDOMAIN means free. The label is sent to the DNS resolver.
func NewEUOrg() Provider { return newDNS("euorg", "euorg", "eu.org") }

// NewDuckDNS checks a free subdomain under duckdns.org. duckdns.org has no
// wildcard DNS, so only claimed names resolve: `{label}.duckdns.org` resolving
// means taken and an NXDOMAIN means free. The label is sent to the DNS resolver.
func NewDuckDNS() Provider { return newDNS("duckdns", "duckdns", "duckdns.org") }

// NewNoIP checks a free hostname under no-ip.org. no-ip.org has no wildcard DNS,
// so only claimed names resolve: `{label}.no-ip.org` resolving means taken and
// an NXDOMAIN means free. The label is sent to the DNS resolver.
func NewNoIP() Provider { return newDNS("noip", "noip", "no-ip.org") }

// NewDynV6 checks a free hostname under dynv6.net. dynv6.net has no wildcard
// DNS, so only claimed names resolve: `{label}.dynv6.net` resolving means taken
// and an NXDOMAIN means free. The label is sent to the DNS resolver.
func NewDynV6() Provider { return newDNS("dynv6", "dynv6", "dynv6.net") }

// NewNotedLol checks a free subdomain under noted.lol. noted.lol has no wildcard
// DNS, so only claimed names resolve: `{label}.noted.lol` resolving means taken
// and an NXDOMAIN means free. The label is sent to the DNS resolver.
func NewNotedLol() Provider { return newDNS("notedlol", "notedlol", "noted.lol") }

// NewIsAGoodDev checks a free subdomain under is-a-good.dev. is-a-good.dev has
// no wildcard DNS, so only claimed names resolve: `{label}.is-a-good.dev`
// resolving means taken and an NXDOMAIN means free. The label is sent to the
// DNS resolver.
func NewIsAGoodDev() Provider { return newDNS("isagooddev", "isagooddev", "is-a-good.dev") }

// NewPagesDev checks a Cloudflare Pages project subdomain under pages.dev.
// pages.dev has no wildcard DNS, so only claimed names resolve:
// `{label}.pages.dev` resolving means taken and an NXDOMAIN means free. The
// label is sent to the DNS resolver.
func NewPagesDev() Provider { return newDNS("pagesdev", "pagesdev", "pages.dev") }

// NewFly checks a Fly.io app hostname under fly.dev. fly.dev has no wildcard
// DNS, so only claimed names resolve: `{label}.fly.dev` resolving means taken
// and an NXDOMAIN means free. The label is sent to the DNS resolver.
func NewFly() Provider { return newDNS("fly", "fly", "fly.dev") }
