package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// Lens checks a handle on Lens Protocol via its GraphQL API. The verdict comes
// from the body, not the HTTP status: an `account(username.localName:)` lookup
// returns a `data.account` object when the handle is taken and a null
// `data.account` when it is free (both answers are a 200). Source:
// docs/platforms.md — POST https://api.lens.xyz/graphql.
//
// The read is keyless but undocumented; unlike most handle sources, creating a
// handle in the Global `lens/` namespace is free, so the source is not paid.
type Lens struct {
	base
	// BaseURL is the GraphQL origin; "https://api.lens.xyz/graphql" when
	// defaulted.
	BaseURL string
	// Client is used for the request; DefaultClient when nil.
	Client *http.Client
}

// NewLens returns a provider bound to api.lens.xyz.
func NewLens() *Lens {
	return &Lens{
		base:    base{id: "lens", platform: "lens", urlFmt: "https://hey.xyz/u/%s"},
		BaseURL: "https://api.lens.xyz/graphql",
		Client:  DefaultClient,
	}
}

// Check POSTs a GraphQL `account(username.localName:)` query. A non-null
// `data.account` object is taken; an explicit null `data.account` is available.
// A missing `data` or `account` field, invalid JSON, or a transport error is
// unknown.
func (l *Lens) Check(ctx context.Context, label string) (Result, error) {
	body, err := json.Marshal(map[string]any{
		"query":     "query ($name: String) { account(request: {username: {localName: $name}}) { username { localName value } } }",
		"variables": map[string]string{"name": label},
	})
	if err != nil {
		return Result{}, err
	}

	resp, err := doPost(ctx, l.Client, l.BaseURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}
	defer drainAndClose(resp)

	var data struct {
		Data *struct {
			Account json.RawMessage `json:"account"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return Result{Status: StatusUnknown}, nil
	}
	if data.Data == nil || len(data.Data.Account) == 0 {
		return Result{Status: StatusUnknown}, nil
	}
	if string(data.Data.Account) == "null" {
		return Result{Status: StatusAvailable}, nil
	}
	return Result{Status: StatusTaken}, nil
}
