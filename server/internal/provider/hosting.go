package provider

// NewVercel checks a project subdomain on Vercel. Vercel serves wildcard DNS
// for vercel.app, so the HTTP status is the signal: `https://{name}.vercel.app/`
// answers 200 when the name is claimed and 404 when it is free. The origin is
// empty because the label is part of the host, not a path segment.
func NewVercel() Provider {
	return newEndpoint("vercel", "vercel",
		"", "https://%s.vercel.app/", "https://%s.vercel.app/", defaultCodes)
}
