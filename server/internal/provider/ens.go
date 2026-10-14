package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// ENS checks whether a name is registered under the Ethereum Name Service
// (`.eth`). The verdict comes from the resolve body, not the HTTP status: the
// api.ensideas.com resolver always answers 200, with a non-null `address` for a
// registered name and `null` for a free one. The metadata.ens.domains avatar
// route 404s both for free names and for registered names without an avatar, so
// it cannot distinguish them.
//
// This leans on a third-party resolver; a follow-up could query a public
// Ethereum RPC against the ENS registry directly.
type ENS struct {
	base
	// BaseURL is the resolver origin; "https://api.ensideas.com" when defaulted.
	BaseURL string
	// Client is used for the request; DefaultClient when nil.
	Client *http.Client
}

// NewENS returns a provider bound to the api.ensideas.com resolver.
func NewENS() *ENS {
	return &ENS{
		base: base{
			id:       "ens",
			platform: "ens",
			nameFmt:  "%s.eth",
			urlFmt:   "https://app.ens.domains/%s.eth",
			paid:     true,
		},
		BaseURL: "https://api.ensideas.com",
		Client:  DefaultClient,
	}
}

// Check resolves "{label}.eth". A 200 whose `address` is non-null is taken; a
// null `address` is available. A non-200, invalid JSON, or a transport error is
// unknown.
func (e *ENS) Check(ctx context.Context, label string) (Result, error) {
	endpoint := e.BaseURL + "/ens/resolve/" + url.PathEscape(label+".eth")
	resp, err := doGet(ctx, e.Client, endpoint)
	if err != nil {
		return Result{}, err
	}
	defer drainAndClose(resp)

	if resp.StatusCode != http.StatusOK {
		return Result{Status: StatusUnknown}, nil
	}
	var data struct {
		Address *string `json:"address"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return Result{Status: StatusUnknown}, nil
	}
	if data.Address == nil || *data.Address == "" {
		return Result{Status: StatusAvailable}, nil
	}
	return Result{Status: StatusTaken}, nil
}
