package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// The providers in this file are the credential-gated ("keyed") sources:
// Discogs and Twitch. Unlike the keyless sources they cannot answer
// without a token, so Default registers them only when provider.Config carries
// their credentials (see keyedSources in provider.go); an absent credential
// omits the source entirely. Every secret travels in a request header or POST
// body and never reaches a Result, a log, or a URL.

// Discogs checks a user on Discogs. Its users endpoint answers 403 without a
// token, 200 for a claimed username and 404 for a free one. The token travels
// in the Authorization header as "Discogs token=<token>".
type Discogs struct {
	base
	// Token is the personal access token (DOMAINBROWSER_DISCOGS_TOKEN).
	Token string
	// BaseURL is the API origin; "https://api.discogs.com" when defaulted.
	BaseURL string
	// Client is used for the request; DefaultClient when nil.
	Client *http.Client
}

// NewDiscogs returns a provider bound to api.discogs.com and authenticated with
// token.
func NewDiscogs(token string) *Discogs {
	return &Discogs{
		base: base{
			id:       "discogs",
			platform: "discogs",
			urlFmt:   "https://www.discogs.com/user/%s",
		},
		Token:   token,
		BaseURL: "https://api.discogs.com",
		Client:  DefaultClient,
	}
}

// Check GETs the user endpoint. A 200 means the user exists, a 404 means the
// name is free, and any other status (including the 403 a missing/invalid token
// returns) is unknown.
func (d *Discogs) Check(ctx context.Context, label string) (Result, error) {
	target := d.BaseURL + "/users/" + url.PathEscape(label)
	resp, err := doGetHeaders(ctx, d.Client, target, map[string]string{
		"Authorization": "Discogs token=" + d.Token,
	})
	if err != nil {
		return Result{}, err
	}
	defer drainAndClose(resp)

	return defaultCodes.classify(resp.StatusCode), nil
}

// twitchTokenMargin refreshes an app access token slightly before it expires,
// so a scan never races the expiry.
const twitchTokenMargin = 60 * time.Second

// Twitch checks a login on Twitch. It authenticates with the client-credentials
// grant: it exchanges the Client-Id/secret for an app access token, caches the
// token until shortly before its expiry, and refreshes it when Helix answers
// 401. The secret travels in the token request body; the client id and bearer
// token travel in Helix headers.
type Twitch struct {
	base
	// ClientID and ClientSecret authorize the app
	// (DOMAINBROWSER_TWITCH_CLIENT_ID, DOMAINBROWSER_TWITCH_CLIENT_SECRET).
	ClientID     string
	ClientSecret string
	// BaseURL is the Helix origin; "https://api.twitch.tv/helix" when defaulted.
	BaseURL string
	// TokenURL is the OAuth token endpoint; the Twitch default when empty.
	TokenURL string
	// Client is used for both requests; DefaultClient when nil.
	Client *http.Client
	// Now stamps the token expiry; time.Now when nil.
	Now func() time.Time

	mu          sync.Mutex
	token       string
	tokenExpiry time.Time
}

// NewTwitch returns a provider bound to Helix and authenticated with the given
// client credentials.
func NewTwitch(clientID, clientSecret string) *Twitch {
	return &Twitch{
		base: base{
			id:       "twitch",
			platform: "twitch",
			urlFmt:   "https://www.twitch.tv/%s",
		},
		ClientID:     clientID,
		ClientSecret: clientSecret,
		BaseURL:      "https://api.twitch.tv/helix",
		TokenURL:     "https://id.twitch.tv/oauth2/token",
		Client:       DefaultClient,
		Now:          time.Now,
	}
}

// Check GETs the users endpoint by login. A 401 refreshes the cached token once
// and retries. A 200 with a non-empty data array is taken, an empty array is
// free; any other status or an invalid body is unknown.
func (t *Twitch) Check(ctx context.Context, label string) (Result, error) {
	status, body, err := t.fetchUsers(ctx, label, false)
	if err != nil {
		return Result{}, err
	}
	if status == http.StatusUnauthorized {
		// The cached token expired early; force a refresh and try once more.
		status, body, err = t.fetchUsers(ctx, label, true)
		if err != nil {
			return Result{}, err
		}
	}
	if status != http.StatusOK {
		return Result{Status: StatusUnknown}, nil
	}

	var data struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return Result{Status: StatusUnknown}, nil
	}
	if len(data.Data) == 0 {
		return Result{Status: StatusAvailable}, nil
	}
	return Result{Status: StatusTaken}, nil
}

// fetchUsers requests the Helix users endpoint with a valid app token and
// returns the status and body. force bypasses the token cache.
func (t *Twitch) fetchUsers(ctx context.Context, label string, force bool) (int, []byte, error) {
	token, err := t.accessToken(ctx, force)
	if err != nil {
		return 0, nil, err
	}
	target := t.BaseURL + "/users?login=" + url.QueryEscape(label)
	resp, err := doGetHeaders(ctx, t.Client, target, map[string]string{
		"Authorization": "Bearer " + token,
		"Client-Id":     t.ClientID,
	})
	if err != nil {
		return 0, nil, err
	}
	defer drainAndClose(resp)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil, nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, body, nil
}

// accessToken returns a cached app access token, fetching a fresh one when the
// cache is empty, expired, or force is set. The client secret stays inside the
// token request body and never reaches a Result or a log.
func (t *Twitch) accessToken(ctx context.Context, force bool) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.Now
	if now == nil {
		now = time.Now
	}
	if !force && t.token != "" && now().Before(t.tokenExpiry) {
		return t.token, nil
	}

	form := url.Values{
		"client_id":     {t.ClientID},
		"client_secret": {t.ClientSecret},
		"grant_type":    {"client_credentials"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	setCommonHeaders(req, defaultAccept)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := t.Client
	if client == nil {
		client = DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer drainAndClose(resp)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("twitch: token request failed with status %d", resp.StatusCode)
	}
	var data struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	if data.AccessToken == "" {
		return "", errors.New("twitch: token response missing access_token")
	}

	ttl := time.Duration(data.ExpiresIn) * time.Second
	if ttl > twitchTokenMargin {
		ttl -= twitchTokenMargin
	}
	t.token = data.AccessToken
	t.tokenExpiry = now().Add(ttl)
	return t.token, nil
}
