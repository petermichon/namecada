package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// PeerTube checks a name on a single PeerTube instance.
// Source: docs/platforms.md — framatube.org shares one local actor namespace
// between /api/v1/accounts/{name} and /api/v1/video-channels/{name}.
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

// accountURL is the public page for an account name, the provider's default.
func (p *PeerTube) accountURL(label string) string {
	return p.URL(label)
}

// channelURL is the public page for a video-channel name. PeerTube serves
// channels under /c/, unlike the /a/ path accounts use.
func (p *PeerTube) channelURL(label string) string {
	format := strings.Replace(p.urlFmt, "/a/", "/c/", 1)
	return fmt.Sprintf(format, url.PathEscape(label))
}

// Check queries both /api/v1/accounts/{label} and /api/v1/video-channels/{label}
// because a PeerTube instance gives accounts and channels one shared local
// actor namespace (the `actor` table is unique on lower(preferredUsername)).
// Either a 200 means the name is taken; only when both answer 404 is it free.
func (p *PeerTube) Check(ctx context.Context, label string) (Result, error) {
	name := url.PathEscape(label)
	accounts := fmt.Sprintf("%s/api/v1/accounts/%s", p.BaseURL, name)
	channels := fmt.Sprintf("%s/api/v1/video-channels/%s", p.BaseURL, name)

	account, err := checkHTTP(ctx, p.Client, accounts, defaultCodes)
	if err != nil {
		return Result{}, err
	}
	if account.Status == StatusTaken {
		account.URL = p.accountURL(label)
		return account, nil
	}

	channel, err := checkHTTP(ctx, p.Client, channels, defaultCodes)
	if err != nil {
		return Result{}, err
	}
	if channel.Status == StatusTaken {
		channel.URL = p.channelURL(label)
		return channel, nil
	}

	if account.Status == StatusAvailable && channel.Status == StatusAvailable {
		return Result{Status: StatusAvailable}, nil
	}
	return Result{Status: StatusUnknown}, nil
}
