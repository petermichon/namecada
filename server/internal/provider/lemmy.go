package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Lemmy checks a username on a single Lemmy instance.
// Source: docs/platforms.md — lemmy.world /api/v3/user?username={name}.
type Lemmy struct {
	base
	// BaseURL is the instance origin the API is called on.
	BaseURL string
	// Client is used for the request; DefaultClient when nil.
	Client *http.Client
}

// NewLemmy returns a provider bound to lemmy.world.
func NewLemmy() *Lemmy {
	return &Lemmy{
		base:    base{id: "lemmy", platform: "lemmy", urlFmt: "https://lemmy.world/u/%s"},
		BaseURL: "https://lemmy.world",
		Client:  DefaultClient,
	}
}

// Check calls /api/v3/user?username={label}: 200 = taken, 404 = available.
func (l *Lemmy) Check(ctx context.Context, label string) (Result, error) {
	endpoint := fmt.Sprintf("%s/api/v3/user?username=%s", l.BaseURL, url.QueryEscape(label))
	return checkHTTP(ctx, l.Client, endpoint, defaultCodes)
}
