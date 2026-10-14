package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// AniList checks a user on AniList via its GraphQL API. The verdict comes from
// the body, not the HTTP status: a `User(name:)` lookup returns a `data.User`
// object when the name is taken and a null `data.User` (with a "Not Found."
// error) when it is free. The free response is a 404, so the body is parsed
// regardless of status. Source: docs/platforms.md — POST https://graphql.anilist.co.
type AniList struct {
	base
	// BaseURL is the GraphQL origin; "https://graphql.anilist.co" when defaulted.
	BaseURL string
	// Client is used for the request; DefaultClient when nil.
	Client *http.Client
}

// NewAniList returns a provider bound to graphql.anilist.co.
func NewAniList() *AniList {
	return &AniList{
		base:    base{id: "anilist", platform: "anilist", urlFmt: "https://anilist.co/user/%s"},
		BaseURL: "https://graphql.anilist.co",
		Client:  DefaultClient,
	}
}

// Check POSTs a GraphQL `User(name:)` query. A non-null `data.User` object is
// taken; an explicit null `data.User` is available. A missing `data` or `User`
// field, invalid JSON, or a transport error is unknown.
func (a *AniList) Check(ctx context.Context, label string) (Result, error) {
	body, err := json.Marshal(map[string]any{
		"query":     "query ($name: String) { User(name: $name) { id } }",
		"variables": map[string]string{"name": label},
	})
	if err != nil {
		return Result{}, err
	}

	resp, err := doPost(ctx, a.Client, a.BaseURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}
	defer drainAndClose(resp)

	var data struct {
		Data *struct {
			User json.RawMessage `json:"User"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return Result{Status: StatusUnknown}, nil
	}
	if data.Data == nil || len(data.Data.User) == 0 {
		return Result{Status: StatusUnknown}, nil
	}
	if string(data.Data.User) == "null" {
		return Result{Status: StatusAvailable}, nil
	}
	return Result{Status: StatusTaken}, nil
}
