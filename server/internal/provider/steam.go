package provider

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Steam checks a custom profile id on Steam. The XML profile endpoint answers
// 200 whether or not the name exists, so the body carries the signal: a taken
// id reports a <steamID64> element, a free one a <error> element. Source:
// docs/platforms.md — GET https://steamcommunity.com/id/{name}?xml=1.
type Steam struct {
	base
	// BaseURL is the profile origin; "https://steamcommunity.com" when defaulted.
	BaseURL string
	// Client is used for the request; DefaultClient when nil.
	Client *http.Client
}

// NewSteam returns a provider bound to steamcommunity.com.
func NewSteam() *Steam {
	return &Steam{
		base:    base{id: "steam", platform: "steam", urlFmt: "https://steamcommunity.com/id/%s"},
		BaseURL: "https://steamcommunity.com",
		Client:  DefaultClient,
	}
}

// Check GETs the XML profile for the custom id. A 200 whose body carries a
// <steamID64> element is taken; a <error> element is free. A non-200, an
// unreadable body, or a body matching neither is unknown.
func (s *Steam) Check(ctx context.Context, label string) (Result, error) {
	target := s.BaseURL + "/id/" + url.PathEscape(label) + "?xml=1"
	resp, err := doGet(ctx, s.Client, target)
	if err != nil {
		return Result{}, err
	}
	defer drainAndClose(resp)

	if resp.StatusCode != http.StatusOK {
		return Result{Status: StatusUnknown}, nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{Status: StatusUnknown}, nil
	}
	text := string(body)
	if strings.Contains(text, "<steamID64>") {
		return Result{Status: StatusTaken}, nil
	}
	if strings.Contains(text, "<error>") {
		return Result{Status: StatusAvailable}, nil
	}
	return Result{Status: StatusUnknown}, nil
}
