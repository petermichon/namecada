package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// Roblox checks a username on Roblox via its usernames lookup. The verdict
// comes from the body, not the HTTP status: a POST for the exact username
// returns a `data` array holding the matching account when the name is taken
// and an empty array when it is free. Source: docs/platforms.md — POST
// https://users.roblox.com/v1/usernames/users.
type Roblox struct {
	base
	// BaseURL is the API origin; "https://users.roblox.com" when defaulted.
	BaseURL string
	// Client is used for the request; DefaultClient when nil.
	Client *http.Client
}

// NewRoblox returns a provider bound to users.roblox.com.
func NewRoblox() *Roblox {
	return &Roblox{
		base: base{
			id:       "roblox",
			platform: "roblox",
			urlFmt:   "https://www.roblox.com/users?username=%s",
		},
		BaseURL: "https://users.roblox.com",
		Client:  DefaultClient,
	}
}

// Check POSTs the username to the usernames lookup. A non-empty `data` array is
// taken, an explicit empty array is available. A missing `data` field, a
// non-200 status, or invalid JSON is unknown.
func (r *Roblox) Check(ctx context.Context, label string) (Result, error) {
	body, err := json.Marshal(map[string]any{
		"usernames":          []string{label},
		"excludeBannedUsers": false,
	})
	if err != nil {
		return Result{}, err
	}

	resp, err := doPost(ctx, r.Client, r.BaseURL+"/v1/usernames/users", "application/json", bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}
	defer drainAndClose(resp)

	if resp.StatusCode != http.StatusOK {
		return Result{Status: StatusUnknown}, nil
	}
	var data struct {
		Data []json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return Result{Status: StatusUnknown}, nil
	}
	if data.Data == nil {
		return Result{Status: StatusUnknown}, nil
	}
	if len(data.Data) == 0 {
		return Result{Status: StatusAvailable}, nil
	}
	return Result{Status: StatusTaken}, nil
}
