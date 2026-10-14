package provider

// platformNames maps a platform id to the human-readable name a result is shown
// under. Providers decide which platforms exist; this registry only labels the
// result the engine streams back, so the web app does not need its own copy.
var platformNames = map[string]string{
	"domains":        "Domains",
	"npm":            "npm",
	"pypi":           "PyPI",
	"crates":         "crates.io",
	"nuget":          "NuGet",
	"rubygems":       "RubyGems",
	"hex":            "Hex",
	"pub":            "pub.dev",
	"homebrew":       "Homebrew",
	"hackage":        "Hackage",
	"clojars":        "Clojars",
	"chocolatey":     "Chocolatey",
	"aur":            "AUR",
	"firefox":        "Firefox Add-ons",
	"jsr":            "JSR",
	"wordpress":      "WordPress.org",
	"openvsx":        "Open VSX",
	"github":         "GitHub",
	"gitlab":         "GitLab",
	"bitbucket":      "Bitbucket",
	"codeberg":       "Codeberg",
	"wikipedia":      "Wikipedia",
	"devto":          "DEV",
	"farcaster":      "Farcaster",
	"minehut":        "Minehut",
	"lichess":        "Lichess",
	"chesscom":       "Chess.com",
	"huggingface":    "Hugging Face",
	"dailymotion":    "Dailymotion",
	"mixcloud":       "Mixcloud",
	"bluesky":        "Bluesky",
	"mastodon":       "Mastodon",
	"opencollective": "Open Collective",
	"lemmy":          "Lemmy",
	"peertube":       "PeerTube",
	"vercel":         "Vercel",
	"netlify":        "Netlify",
	"surge":          "Surge",
	"firebase":       "Firebase Hosting",
	"neocities":      "Neocities",
	"render":         "Render",
	"railway":        "Railway",
	"heroku":         "Heroku",
	"denodeploy":     "Deno Deploy",
	"readthedocs":    "Read the Docs",
	"euorg":          "eu.org",
	"duckdns":        "DuckDNS",
	"noip":           "No-IP",
	"dynv6":          "dynv6",
	"notedlol":       "noted.lol",
	"isagooddev":     "is-a-good.dev",
}

// PlatformName returns the display name for a platform id. An unknown id is
// returned unchanged so a newly added provider is still labelled.
func PlatformName(platform string) string {
	if name, ok := platformNames[platform]; ok {
		return name
	}
	return platform
}
