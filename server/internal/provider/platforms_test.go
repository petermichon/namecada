package provider

import "testing"

func TestPlatformName(t *testing.T) {
	cases := map[string]string{
		"domains":        "Domains",
		"pypi":           "PyPI",
		"crates":         "crates.io",
		"firefox":        "Firefox Add-ons",
		"chesscom":       "Chess.com",
		"huggingface":    "Hugging Face",
		"opencollective": "Open Collective",
	}
	for platform, want := range cases {
		if got := PlatformName(platform); got != want {
			t.Errorf("PlatformName(%q) = %q, want %q", platform, got, want)
		}
	}
}

func TestPlatformNameFallsBackToID(t *testing.T) {
	if got := PlatformName("brand-new"); got != "brand-new" {
		t.Errorf("PlatformName(unknown) = %q, want the id unchanged", got)
	}
}

func TestDefaultPlatformsHaveNames(t *testing.T) {
	for _, p := range Default().All() {
		platform := p.Platform()
		if _, ok := platformNames[platform]; !ok {
			t.Errorf("provider %q has platform %q with no display name", p.ID(), platform)
		}
	}
}
