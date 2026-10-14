package provider

// newHostEndpoint builds a host-status endpoint whose probe hits an HTML route.
// The whole family overrides the default JSON Accept with htmlAccept because a
// claimed app can answer 406 Not Acceptable to application/json. The origin is
// empty: the label is part of the host, not a path segment.
func newHostEndpoint(id, platform, path, pageURL string, codes codeMap) *endpoint {
	e := newEndpoint(id, platform, "", path, pageURL, codes)
	e.accept = htmlAccept
	return e
}

// NewVercel checks a project subdomain on Vercel. Vercel serves wildcard DNS
// for vercel.app, so the HTTP status is the signal: `https://{name}.vercel.app/`
// answers 200 when the name is claimed and 404 when it is free. The origin is
// empty because the label is part of the host, not a path segment.
func NewVercel() Provider {
	return newHostEndpoint("vercel", "vercel",
		"https://%s.vercel.app/", "https://%s.vercel.app/", defaultCodes)
}

// NewNetlify checks a site subdomain on Netlify. Netlify serves wildcard DNS
// for netlify.app, so the HTTP status is the signal: `https://{name}.netlify.app/`
// answers 200 when the name is claimed and 404 when it is free. The origin is
// empty because the label is part of the host, not a path segment.
func NewNetlify() Provider {
	return newHostEndpoint("netlify", "netlify",
		"https://%s.netlify.app/", "https://%s.netlify.app/", defaultCodes)
}

// NewSurge checks a project subdomain on Surge. Surge serves wildcard DNS for
// surge.sh, so the HTTP status is the signal: `https://{name}.surge.sh/`
// answers 200 when the name is claimed and 404 when it is free. The origin is
// empty because the label is part of the host, not a path segment.
func NewSurge() Provider {
	return newHostEndpoint("surge", "surge",
		"https://%s.surge.sh/", "https://%s.surge.sh/", defaultCodes)
}

// NewFirebase checks a Hosting subdomain on Firebase. Firebase serves wildcard
// DNS for web.app, so the HTTP status is the signal:
// `https://{name}.web.app/` answers 200 when the name is claimed and 404 when it
// is free. The name is the globally-unique Firebase project id. The origin is
// empty because the label is part of the host, not a path segment.
func NewFirebase() Provider {
	return newHostEndpoint("firebase", "firebase",
		"https://%s.web.app/", "https://%s.web.app/", defaultCodes)
}

// NewNeocities checks a site subdomain on Neocities. Neocities serves wildcard
// DNS for neocities.org, so the HTTP status is the signal:
// `https://{name}.neocities.org/` answers 200 when the name is claimed and 404
// when it is free. The origin is empty because the label is part of the host,
// not a path segment.
func NewNeocities() Provider {
	return newHostEndpoint("neocities", "neocities",
		"https://%s.neocities.org/", "https://%s.neocities.org/", defaultCodes)
}

// NewRender checks a service subdomain on Render. Render serves wildcard DNS
// for onrender.com, so the HTTP status is the signal:
// `https://{name}.onrender.com/` answers 200 when the name is claimed and 404
// when it is free. The origin is empty because the label is part of the host,
// not a path segment. A free-tier service that has spun down answers 503, which
// is classified unknown and never cached, so it self-heals.
func NewRender() Provider {
	return newHostEndpoint("render", "render",
		"https://%s.onrender.com/", "https://%s.onrender.com/", defaultCodes)
}

// NewRailway checks a service subdomain on Railway. Railway serves wildcard DNS
// for up.railway.app, so the HTTP status is the signal:
// `https://{name}.up.railway.app/` answers 200 when the name is claimed and 404
// when it is free. Service names are user-chosen; on a collision Railway appends
// a suffix, so a free name still answers 404 for the exact label. The origin is
// empty because the label is part of the host, not a path segment. Some claimed
// apps answer 406 to a JSON Accept, so the route overrides it with htmlAccept.
func NewRailway() Provider {
	return newHostEndpoint("railway", "railway",
		"https://%s.up.railway.app/", "https://%s.up.railway.app/", defaultCodes)
}

// NewHeroku checks an app subdomain on Heroku. Heroku serves wildcard DNS for
// herokuapp.com, so the HTTP status is the signal:
// `https://{name}.herokuapp.com/` answers 200 when the name is claimed and 404
// when it is free. App names are user-chosen, and a claimed app may answer 302
// before the client follows it to a 200. The origin is empty because the label
// is part of the host, not a path segment. Heroku has been paid-only since
// November 2022, so the source is marked paid; it is still surfaced.
func NewHeroku() Provider {
	e := newHostEndpoint("heroku", "heroku",
		"https://%s.herokuapp.com/", "https://%s.herokuapp.com/", defaultCodes)
	e.paid = true
	return e
}

// NewDenoDeploy checks a project subdomain on Deno Deploy. Deno Deploy serves
// wildcard DNS for deno.dev, so the HTTP status is the signal:
// `https://{name}.deno.dev/` answers 404 when the name is free. A claimed name
// can answer 301 before redirecting to the project's own custom domain (for
// example `fresh.deno.dev` → `usefresh.dev`), and DefaultClient follows
// redirects, so the final 200 is what classify sees. Project names are
// user-chosen. The origin is empty because the label is part of the host, not a
// path segment.
func NewDenoDeploy() Provider {
	return newHostEndpoint("denodeploy", "denodeploy",
		"https://%s.deno.dev/", "https://%s.deno.dev/", defaultCodes)
}

// NewGumroad checks a creator storefront subdomain on Gumroad. Gumroad serves
// wildcard DNS for gumroad.com, so the HTTP status is the signal:
// `https://{name}.gumroad.com/` answers 200 when the name is claimed and 404
// when it is free. Storefront names are user-chosen. The origin is empty because
// the label is part of the host, not a path segment.
func NewGumroad() Provider {
	return newHostEndpoint("gumroad", "gumroad",
		"https://%s.gumroad.com/", "https://%s.gumroad.com/", defaultCodes)
}

// NewReadTheDocs checks a project subdomain on Read the Docs. Read the Docs
// serves wildcard DNS for readthedocs.io, so the HTTP status is the signal:
// `https://{name}.readthedocs.io/` answers 302 to the project's latest docs
// when the name is claimed and 404 when it is free. DefaultClient follows the
// redirect, so the final 200 is what classify sees. Project slugs are
// user-chosen. The origin is empty because the label is part of the host, not a
// path segment.
func NewReadTheDocs() Provider {
	return newHostEndpoint("readthedocs", "readthedocs",
		"https://%s.readthedocs.io/", "https://%s.readthedocs.io/", defaultCodes)
}
