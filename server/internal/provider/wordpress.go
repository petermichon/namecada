package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

// WordPress checks whether a plugin slug is taken on WordPress.org. Unlike the
// status-only registries it inspects the 404 body: WordPress reports closed
// plugins as 404 with {"closed":true}, so a plain 404 would falsely read as
// available. Source: docs/platforms.md — api.wordpress.org/plugins/info/1.2.
type WordPress struct {
	base
	// BaseURL is the API origin; "https://api.wordpress.org" when defaulted.
	BaseURL string
	// Client is used for the request; DefaultClient when nil.
	Client *http.Client
}

// NewWordPress returns a provider bound to api.wordpress.org.
func NewWordPress() *WordPress {
	return &WordPress{
		base:    base{id: "wordpress", platform: "wordpress", urlFmt: "https://wordpress.org/plugins/%s/"},
		BaseURL: "https://api.wordpress.org",
		Client:  DefaultClient,
	}
}

// Check queries the plugin information API. 200 means the plugin exists; a 404
// with {"closed":true} means it is taken but closed; any other 404 is free.
func (w *WordPress) Check(ctx context.Context, label string) (Result, error) {
	query := url.Values{}
	query.Set("action", "plugin_information")
	query.Set("request[slug]", label)
	endpoint := w.BaseURL + "/plugins/info/1.2/?" + query.Encode()

	client := w.Client
	if client == nil {
		client = DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Result{}, err
	}
	setCommonHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		return Result{Status: StatusTaken}, nil
	case http.StatusNotFound:
		var data struct {
			Closed bool `json:"closed"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&data)
		if data.Closed {
			return Result{Status: StatusTaken}, nil
		}
		return Result{Status: StatusAvailable}, nil
	default:
		return Result{Status: StatusUnknown}, nil
	}
}
