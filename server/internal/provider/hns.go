package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// HNS checks whether a name is registered under Handshake (HNS). The verdict
// comes from the body, not the HTTP status: the hnsscan.com explorer answers
// 200 for a known name with `info.state` of "CLOSED", "AVAILABLE", or
// "UNKNOWN". "CLOSED" is a name whose auction has closed and is registered;
// "AVAILABLE" is an expired or never-claimed name; "UNKNOWN" is a name the
// chain has never seen. The historical `info.registered` flag stays true even
// for an expired name (e.g. `hnscontrol`), so it must not be used alone. A
// reserved name (e.g. `example`) answers 502. Any other status, a missing
// `info`, or invalid JSON degrades to unknown.
//
// This leans on a third-party explorer with an undocumented API.
// Source: docs/platforms.md — GET https://hnsscan.com/api/names/{name}/details.
type HNS struct {
	base
	// BaseURL is the explorer origin; "https://hnsscan.com" when defaulted.
	BaseURL string
	// Client is used for the request; DefaultClient when nil.
	Client *http.Client
}

// NewHNS returns a provider bound to the hnsscan.com explorer.
func NewHNS() *HNS {
	return &HNS{
		base: base{
			id:       "hns",
			platform: "hns",
			urlFmt:   "https://hnsscan.com/names/%s",
			paid:     true,
		},
		BaseURL: "https://hnsscan.com",
		Client:  DefaultClient,
	}
}

// Check reads the name's `info.state`. "CLOSED" is taken; "AVAILABLE" and
// "UNKNOWN" are available. A non-200, a missing `info`, an unrecognised state,
// or invalid JSON is unknown.
func (h *HNS) Check(ctx context.Context, label string) (Result, error) {
	endpoint := h.BaseURL + "/api/names/" + url.PathEscape(label) + "/details"
	resp, err := doGet(ctx, h.Client, endpoint)
	if err != nil {
		return Result{}, err
	}
	defer drainAndClose(resp)

	if resp.StatusCode != http.StatusOK {
		return Result{Status: StatusUnknown}, nil
	}
	var data struct {
		Info *struct {
			State string `json:"state"`
		} `json:"info"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return Result{Status: StatusUnknown}, nil
	}
	if data.Info == nil {
		return Result{Status: StatusUnknown}, nil
	}
	switch data.Info.State {
	case "CLOSED":
		return Result{Status: StatusTaken}, nil
	case "AVAILABLE", "UNKNOWN":
		return Result{Status: StatusAvailable}, nil
	default:
		return Result{Status: StatusUnknown}, nil
	}
}
