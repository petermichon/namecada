package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// SNS checks whether a name is registered under the Solana Name Service
// (`.sol`). The verdict comes from the resolve body, not the HTTP status: the
// sdk-proxy.sns.id resolver always answers 200, with `{"s":"ok","result":
// "<pubkey>"}` for a registered name and `{"s":"error","result":"Domain not
// found"}` for a free one.
//
// This leans on a third-party SDK proxy with an undocumented wire shape; any
// response that is not one of the two known envelopes degrades to unknown.
type SNS struct {
	base
	// BaseURL is the resolver origin; "https://sdk-proxy.sns.id" when defaulted.
	BaseURL string
	// Client is used for the request; DefaultClient when nil.
	Client *http.Client
}

// NewSNS returns a provider bound to the sdk-proxy.sns.id resolver.
func NewSNS() *SNS {
	return &SNS{
		base: base{
			id:       "sns",
			platform: "sns",
			nameFmt:  "%s.sol",
			urlFmt:   "https://www.sns.id/domain/%s.sol",
			paid:     true,
		},
		BaseURL: "https://sdk-proxy.sns.id",
		Client:  DefaultClient,
	}
}

// Check resolves "{label}.sol". A 200 whose `s` is "ok" with a non-empty
// `result` is taken; `s` "error" is available. A non-200, invalid JSON, or any
// other envelope is unknown.
func (s *SNS) Check(ctx context.Context, label string) (Result, error) {
	endpoint := s.BaseURL + "/resolve/" + url.PathEscape(label+".sol")
	resp, err := doGet(ctx, s.Client, endpoint)
	if err != nil {
		return Result{}, err
	}
	defer drainAndClose(resp)

	if resp.StatusCode != http.StatusOK {
		return Result{Status: StatusUnknown}, nil
	}
	var data struct {
		S      string `json:"s"`
		Result string `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return Result{Status: StatusUnknown}, nil
	}
	switch data.S {
	case "ok":
		if data.Result == "" {
			return Result{Status: StatusUnknown}, nil
		}
		return Result{Status: StatusTaken}, nil
	case "error":
		return Result{Status: StatusAvailable}, nil
	default:
		return Result{Status: StatusUnknown}, nil
	}
}
