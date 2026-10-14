package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// PeerTube checks an account on a single PeerTube instance.
// Source: docs/platforms.md — framatube.org /api/v1/accounts/{name}.
type PeerTube struct {
	base
	// BaseURL is the instance origin the API is called on.
	BaseURL string
	// Client is used for the request; DefaultClient when nil.
	Client *http.Client
}

// NewPeerTube returns a provider bound to framatube.org.
func NewPeerTube() *PeerTube {
	return &PeerTube{
		base:    base{id: "peertube", platform: "peertube", urlFmt: "https://framatube.org/a/%s"},
		BaseURL: "https://framatube.org",
		Client:  DefaultClient,
	}
}

// Check calls /api/v1/accounts/{label}: 200 = taken, 404 = available.
func (p *PeerTube) Check(ctx context.Context, label string) (Result, error) {
	endpoint := fmt.Sprintf("%s/api/v1/accounts/%s", p.BaseURL, url.PathEscape(label))
	return checkHTTP(ctx, p.Client, endpoint, defaultCodes)
}
