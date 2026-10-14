package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// Namecoin checks whether a name is registered under Namecoin (`.bit`). The
// verdict comes from the body, not the HTTP status: the explorer answers 404
// for a name that has never been registered, and 200 for a known one with an
// `expired` flag. A 200 with `expired:true` is a name whose registration has
// lapsed and is therefore claimable; a 200 with `expired:false` is a live
// registration. The name lives in Namecoin's `d/` namespace, the one the `.bit`
// resolver serves.
//
// This leans on a third-party explorer with an undocumented API.
// Source: docs/platforms.md — GET https://explorer.namecoin.co/api/name/d%2F{name}.
type Namecoin struct {
	base
	// BaseURL is the explorer origin; "https://explorer.namecoin.co" when
	// defaulted.
	BaseURL string
	// Client is used for the request; DefaultClient when nil.
	Client *http.Client
}

// NewNamecoin returns a provider bound to the explorer.namecoin.co explorer.
func NewNamecoin() *Namecoin {
	return &Namecoin{
		base: base{
			id:       "namecoin",
			platform: "namecoin",
			urlFmt:   "https://explorer.namecoin.co/name/d/%s",
			paid:     true,
		},
		BaseURL: "https://explorer.namecoin.co",
		Client:  DefaultClient,
	}
}

// Check reads the name's `expired` flag. A 404 is available; a 200 with
// `expired:true` is available and `expired:false` is taken. Any other status, a
// missing `expired`, or invalid JSON is unknown.
func (n *Namecoin) Check(ctx context.Context, label string) (Result, error) {
	endpoint := n.BaseURL + "/api/name/d%2F" + url.PathEscape(label)
	resp, err := doGet(ctx, n.Client, endpoint)
	if err != nil {
		return Result{}, err
	}
	defer drainAndClose(resp)

	if resp.StatusCode == http.StatusNotFound {
		return Result{Status: StatusAvailable}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return Result{Status: StatusUnknown}, nil
	}
	var data struct {
		Expired *bool `json:"expired"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return Result{Status: StatusUnknown}, nil
	}
	if data.Expired == nil {
		return Result{Status: StatusUnknown}, nil
	}
	if *data.Expired {
		return Result{Status: StatusAvailable}, nil
	}
	return Result{Status: StatusTaken}, nil
}
