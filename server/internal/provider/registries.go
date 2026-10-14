package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// endpoint is a package-registry provider whose presence signal is the HTTP
// status of a single JSON endpoint. It mirrors the frontend checkHttp helper
// and covers every registry that needs no response-body parsing.
type endpoint struct {
	id       string
	platform string
	// baseURL is the origin the API lives on, without a trailing slash.
	baseURL string
	// path is a format string with one %s for the escaped label.
	path string
	// pageURL is a format string with one %s for the escaped label.
	pageURL string
	codes   codeMap
	client  *http.Client
	// accept overrides the request Accept header; empty keeps defaultAccept.
	accept string
	// paid marks a source that charges money to register; registries are free.
	paid bool
}

func newEndpoint(id, platform, baseURL, path, pageURL string, codes codeMap) *endpoint {
	return &endpoint{
		id:       id,
		platform: platform,
		baseURL:  baseURL,
		path:     path,
		pageURL:  pageURL,
		codes:    codes,
		client:   DefaultClient,
		accept:   defaultAccept,
	}
}

func (e *endpoint) ID() string               { return e.id }
func (e *endpoint) Platform() string         { return e.platform }
func (e *endpoint) Paid() bool               { return e.paid }
func (e *endpoint) Name(label string) string { return label }

func (e *endpoint) URL(label string) string {
	return fmt.Sprintf(e.pageURL, url.PathEscape(label))
}

// Check calls baseURL + path with the escaped label and classifies the status
// through the endpoint's code table.
func (e *endpoint) Check(ctx context.Context, label string) (Result, error) {
	target := e.baseURL + fmt.Sprintf(e.path, url.PathEscape(label))
	return checkHTTPAccept(ctx, e.client, target, e.codes, e.accept)
}

// NewNPM checks a package on the npm registry.
func NewNPM() Provider {
	return newEndpoint("npm", "npm",
		"https://registry.npmjs.org", "/%s", "https://www.npmjs.com/package/%s", defaultCodes)
}

// NewPyPI checks a project on PyPI.
func NewPyPI() Provider {
	return newEndpoint("pypi", "pypi",
		"https://pypi.org", "/pypi/%s/json", "https://pypi.org/project/%s/", defaultCodes)
}

// NewCrates checks a crate on crates.io.
func NewCrates() Provider {
	return newEndpoint("crates", "crates",
		"https://crates.io", "/api/v1/crates/%s", "https://crates.io/crates/%s", defaultCodes)
}

// NewNuGet checks a package on NuGet.
func NewNuGet() Provider {
	return newEndpoint("nuget", "nuget",
		"https://api.nuget.org", "/v3-flatcontainer/%s/index.json",
		"https://www.nuget.org/packages/%s", defaultCodes)
}

// NewRubyGems checks a gem on RubyGems. A 200 means the gem exists and a 404
// means the name is free.
func NewRubyGems() Provider {
	return newEndpoint("rubygems", "rubygems",
		"https://rubygems.org", "/api/v1/gems/%s.json",
		"https://rubygems.org/gems/%s", defaultCodes)
}

// NewHex checks a package on Hex (Elixir). Its minimum length is declared in
// the known table.
func NewHex() Provider {
	return newEndpoint("hex", "hex",
		"https://hex.pm", "/api/packages/%s", "https://hex.pm/packages/%s", defaultCodes)
}

// NewHexUser checks a user on Hex. It shares the "hex" platform with NewHex.
func NewHexUser() Provider {
	return newEndpoint("hex-user", "hex",
		"https://hex.pm", "/api/users/%s", "https://hex.pm/users/%s", defaultCodes)
}

// NewPub checks a package on pub.dev (Dart).
func NewPub() Provider {
	return newEndpoint("pub", "pub",
		"https://pub.dev", "/api/packages/%s", "https://pub.dev/packages/%s", defaultCodes)
}

// NewHomebrew checks a formula on Homebrew.
func NewHomebrew() Provider {
	return newEndpoint("homebrew", "homebrew",
		"https://formulae.brew.sh", "/api/formula/%s.json",
		"https://formulae.brew.sh/formula/%s", defaultCodes)
}

// NewHackage checks a package on Hackage (Haskell). It serves HTML, and the
// bare package name resolves, so the HTTP status alone is the signal: a 200
// means the package exists and a 404 means the name is free.
func NewHackage() Provider {
	return newEndpoint("hackage", "hackage",
		"https://hackage.haskell.org", "/package/%s",
		"https://hackage.haskell.org/package/%s", defaultCodes)
}

// NewClojars checks an artifact on Clojars. The API answers 200 for a taken
// name and 404 for a free one, and a single segment resolves on the public page.
func NewClojars() Provider {
	return newEndpoint("clojars", "clojars",
		"https://clojars.org", "/api/artifacts/%s",
		"https://clojars.org/%s", defaultCodes)
}

// NewAUR checks a package on the Arch User Repository. The RPC endpoint always
// answers 200, so this uses the package web page, where a 200 means the package
// exists and a 404 means the name is free.
func NewAUR() Provider {
	return newEndpoint("aur", "aur",
		"https://aur.archlinux.org", "/packages/%s",
		"https://aur.archlinux.org/packages/%s", defaultCodes)
}

// NewFirefox checks an add-on on Firefox Add-ons (AMO). A 401 is treated as
// taken alongside 200: the API returns 401 for add-ons that hide their
// existence from unauthenticated callers.
func NewFirefox() Provider {
	return newEndpoint("firefox", "firefox",
		"https://addons.mozilla.org", "/api/v5/addons/addon/%s/",
		"https://addons.mozilla.org/firefox/addon/%s/",
		codeMap{available: []int{http.StatusNotFound}, taken: []int{http.StatusOK, http.StatusUnauthorized}})
}

// NewJSR checks a scope on JSR. A 400 is treated as taken alongside 200: the
// scope API rejects invalid-but-taken names with a 400 rather than a 200.
func NewJSR() Provider {
	return newEndpoint("jsr", "jsr",
		"https://jsr.io", "/api/scopes/%s", "https://jsr.io/@%s",
		codeMap{available: []int{http.StatusNotFound}, taken: []int{http.StatusOK, http.StatusBadRequest}})
}

// NewOpenVSX checks a namespace on Open VSX.
func NewOpenVSX() Provider {
	return newEndpoint("openvsx", "openvsx",
		"https://open-vsx.org", "/api/%s", "https://open-vsx.org/namespace/%s", defaultCodes)
}
