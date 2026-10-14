package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// The providers in this file cover the account/handle sources ported from the
// frontend providers.ts (wikipedia, devto, farcaster, minehut, lichess,
// chesscom, huggingface, dailymotion, mixcloud, bluesky, mastodon,
// opencollective). Status-only sources reuse the endpoint helper; the rest
// inspect the response body.

// NewFarcaster checks a username on Farcaster via the fname registry.
func NewFarcaster() Provider {
	return newEndpoint("farcaster", "farcaster",
		"https://fnames.farcaster.xyz", "/transfers/current?name=%s",
		"https://farcaster.xyz/%s", defaultCodes)
}

// NewLichess checks a user on Lichess. Its minimum length (2) is declared in the
// known table.
func NewLichess() Provider {
	return newEndpoint("lichess", "lichess",
		"https://lichess.org", "/api/user/%s", "https://lichess.org/@/%s", defaultCodes)
}

// NewChesscom checks a player on Chess.com. Its minimum length (3) is declared
// in the known table.
func NewChesscom() Provider {
	return newEndpoint("chesscom", "chesscom",
		"https://api.chess.com", "/pub/player/%s", "https://www.chess.com/member/%s", defaultCodes)
}

// NewHuggingFace checks a user on Hugging Face. Its minimum length (2) is
// declared in the known table.
func NewHuggingFace() Provider {
	return newEndpoint("huggingface", "huggingface",
		"https://huggingface.co", "/api/users/%s/overview", "https://huggingface.co/%s",
		defaultCodes)
}

// NewDailymotion checks a user on Dailymotion. Its minimum length (3) is
// declared in the known table.
func NewDailymotion() Provider {
	return newEndpoint("dailymotion", "dailymotion",
		"https://api.dailymotion.com", "/user/%s", "https://www.dailymotion.com/%s",
		defaultCodes)
}

// NewMixcloud checks a user on Mixcloud.
func NewMixcloud() Provider {
	return newEndpoint("mixcloud", "mixcloud",
		"https://api.mixcloud.com", "/%s/", "https://www.mixcloud.com/%s/", defaultCodes)
}

// NewMastodon checks an account on mastodon.social. The account lookup API
// flattens suspended accounts to 404, so it probes the profile instead: 200 and
// 410 Gone both mean the username is taken, 404 is free. The profile route
// serves HTML, so it overrides the default JSON Accept; sending
// application/json makes the route answer 401.
func NewMastodon() Provider {
	e := newEndpoint("mastodon", "mastodon",
		"https://mastodon.social", "/@%s", "https://mastodon.social/@%s",
		codeMap{
			available: []int{http.StatusNotFound},
			taken:     []int{http.StatusOK, http.StatusGone},
		})
	e.accept = "text/html"
	return e
}

// Wikipedia checks a global account on Wikimedia. Unlike the status-only
// sources it reads the body: globaluserinfo carries a "missing" key when the
// name is free. Source: docs/platforms.md — meta.wikimedia.org/w/api.php.
type Wikipedia struct {
	base
	// BaseURL is the API origin; "https://meta.wikimedia.org" when defaulted.
	BaseURL string
	// Client is used for the request; DefaultClient when nil.
	Client *http.Client
}

// NewWikipedia returns a provider bound to meta.wikimedia.org.
func NewWikipedia() *Wikipedia {
	return &Wikipedia{
		base:    base{id: "wikipedia", platform: "wikipedia", urlFmt: "https://en.wikipedia.org/wiki/Special:CentralAuth/%s"},
		BaseURL: "https://meta.wikimedia.org",
		Client:  DefaultClient,
	}
}

// Check queries the globaluserinfo meta API. A 200 whose globaluserinfo object
// carries a "missing" key is free; any other object is taken. A non-200, a
// missing/null body, or invalid JSON is unknown.
func (w *Wikipedia) Check(ctx context.Context, label string) (Result, error) {
	query := url.Values{}
	query.Set("action", "query")
	query.Set("meta", "globaluserinfo")
	query.Set("guiuser", label)
	query.Set("format", "json")
	query.Set("origin", "*")
	endpoint := w.BaseURL + "/w/api.php?" + query.Encode()

	resp, err := doGet(ctx, w.Client, endpoint)
	if err != nil {
		return Result{}, err
	}
	defer drainAndClose(resp)

	if resp.StatusCode != http.StatusOK {
		return Result{Status: StatusUnknown}, nil
	}
	var data struct {
		Query *struct {
			GlobalUserInfo map[string]json.RawMessage `json:"globaluserinfo"`
		} `json:"query"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return Result{Status: StatusUnknown}, nil
	}
	if data.Query == nil || data.Query.GlobalUserInfo == nil {
		return Result{Status: StatusUnknown}, nil
	}
	if _, ok := data.Query.GlobalUserInfo["missing"]; ok {
		return Result{Status: StatusAvailable}, nil
	}
	return Result{Status: StatusTaken}, nil
}

// DevTo checks a user or organisation on DEV (Forem). It queries the users
// endpoint first and, only when that reports the name free, falls back to the
// organisations endpoint. Its minimum length (2) is declared in the known table.
type DevTo struct {
	base
	// BaseURL is the API origin; "https://dev.to" when defaulted.
	BaseURL string
	// Client is used for the request; DefaultClient when nil.
	Client *http.Client
}

// NewDevTo returns a provider bound to dev.to.
func NewDevTo() *DevTo {
	return &DevTo{
		base:    base{id: "devto", platform: "devto", urlFmt: "https://dev.to/%s"},
		BaseURL: "https://dev.to",
		Client:  DefaultClient,
	}
}

// Check queries the username endpoint; anything other than available is
// returned as-is. Only an available user falls back to the organisation
// endpoint, mirroring the frontend devto check.
func (d *DevTo) Check(ctx context.Context, label string) (Result, error) {
	userURL := d.BaseURL + "/api/users/by_username?url=" + url.QueryEscape(label)
	user, err := checkHTTP(ctx, d.Client, userURL, defaultCodes)
	if err != nil {
		return Result{}, err
	}
	if user.Status != StatusAvailable {
		return user, nil
	}
	orgURL := d.BaseURL + "/api/organizations/" + url.PathEscape(label)
	return checkHTTP(ctx, d.Client, orgURL, defaultCodes)
}

// Minehut checks a Minecraft server name. The API answers 200 with a "server"
// field that is null when the name is free. Its minimum length (4) is declared
// in the known table.
type Minehut struct {
	base
	// BaseURL is the API origin; "https://api.minehut.com" when defaulted.
	BaseURL string
	// Client is used for the request; DefaultClient when nil.
	Client *http.Client
}

// NewMinehut returns a provider bound to api.minehut.com.
func NewMinehut() *Minehut {
	return &Minehut{
		base:    base{id: "minehut", platform: "minehut", urlFmt: "https://www.minehut.com/servers/%s"},
		BaseURL: "https://api.minehut.com",
		Client:  DefaultClient,
	}
}

// Check calls the by-name server endpoint. On a 200 an explicit "server": null
// is free, a present object is taken, and a missing field or invalid body is
// unknown.
func (m *Minehut) Check(ctx context.Context, label string) (Result, error) {
	endpoint := m.BaseURL + "/server/" + url.PathEscape(label) + "?byName=true"
	resp, err := doGet(ctx, m.Client, endpoint)
	if err != nil {
		return Result{}, err
	}
	defer drainAndClose(resp)

	if resp.StatusCode != http.StatusOK {
		return Result{Status: StatusUnknown}, nil
	}
	var data struct {
		Server json.RawMessage `json:"server"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return Result{Status: StatusUnknown}, nil
	}
	if len(data.Server) == 0 {
		return Result{Status: StatusUnknown}, nil
	}
	if string(data.Server) == "null" {
		return Result{Status: StatusAvailable}, nil
	}
	return Result{Status: StatusTaken}, nil
}

// Bluesky checks a handle on bsky.social. resolveHandle cannot see
// reserved/deactivated/over-length handles, so it asks the signup availability
// check instead: a "resultAvailable" $type is free, anything else taken. Its
// minimum length (3) is declared in the known table.
type Bluesky struct {
	base
	// BaseURL is the API origin; "https://bsky.social" when defaulted.
	BaseURL string
	// Client is used for the request; DefaultClient when nil.
	Client *http.Client
}

// NewBluesky returns a provider bound to bsky.social.
func NewBluesky() *Bluesky {
	return &Bluesky{
		base: base{
			id:       "bluesky",
			platform: "bluesky",
			nameFmt:  "%s.bsky.social",
			urlFmt:   "https://bsky.app/profile/%s.bsky.social",
		},
		BaseURL: "https://bsky.social",
		Client:  DefaultClient,
	}
}

// Check calls checkHandleAvailability for "{label}.bsky.social". A 200 with a
// string $type ending in "#resultAvailable" is free; any other type is taken.
// A non-200, a missing/non-string type, or invalid JSON is unknown.
func (b *Bluesky) Check(ctx context.Context, label string) (Result, error) {
	handle := label + ".bsky.social"
	endpoint := b.BaseURL + "/xrpc/com.atproto.temp.checkHandleAvailability?handle=" +
		url.QueryEscape(handle)
	resp, err := doGet(ctx, b.Client, endpoint)
	if err != nil {
		return Result{}, err
	}
	defer drainAndClose(resp)

	if resp.StatusCode != http.StatusOK {
		return Result{Status: StatusUnknown}, nil
	}
	var data struct {
		Result struct {
			Type string `json:"$type"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return Result{Status: StatusUnknown}, nil
	}
	if data.Result.Type == "" {
		return Result{Status: StatusUnknown}, nil
	}
	if strings.HasSuffix(data.Result.Type, "#resultAvailable") {
		return Result{Status: StatusAvailable}, nil
	}
	return Result{Status: StatusTaken}, nil
}

// OpenCollective checks a collective, user or organisation. Its CDN can cache a
// transient 5xx per Origin for up to an hour, so an unknown first response is
// retried once with a cache-busting parameter.
type OpenCollective struct {
	base
	// BaseURL is the API origin; "https://opencollective.com" when defaulted.
	BaseURL string
	// Client is used for the request; DefaultClient when nil.
	Client *http.Client
	// Now stamps the cache-busting retry; time.Now when nil.
	Now func() time.Time
}

// NewOpenCollective returns a provider bound to opencollective.com.
func NewOpenCollective() *OpenCollective {
	return &OpenCollective{
		base:    base{id: "opencollective", platform: "opencollective", urlFmt: "https://opencollective.com/%s"},
		BaseURL: "https://opencollective.com",
		Client:  DefaultClient,
		Now:     time.Now,
	}
}

// Check GETs "{label}.json"; a taken/available result is returned directly. An
// unknown result is retried once with a cache-busting "_" query parameter to
// bypass a poisoned CDN entry.
func (o *OpenCollective) Check(ctx context.Context, label string) (Result, error) {
	target := o.BaseURL + "/" + url.PathEscape(label) + ".json"
	got, err := checkHTTP(ctx, o.Client, target, defaultCodes)
	if err != nil {
		return Result{}, err
	}
	if got.Status != StatusUnknown {
		return got, nil
	}

	now := o.Now
	if now == nil {
		now = time.Now
	}
	bust := target + "?_=" + strconv.FormatInt(now().UnixMilli(), 10)
	return checkHTTP(ctx, o.Client, bust, defaultCodes)
}
