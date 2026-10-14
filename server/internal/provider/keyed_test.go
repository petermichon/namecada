package provider

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// TestDiscogsCheck pins the status mapping and that the token travels in the
// Authorization header, never in the URL.
func TestDiscogsCheck(t *testing.T) {
	tests := []struct {
		name string
		code int
		want Status
	}{
		{"taken", http.StatusOK, StatusTaken},
		{"available", http.StatusNotFound, StatusAvailable},
		{"keyless route forbidden", http.StatusForbidden, StatusUnknown},
		{"server error", http.StatusInternalServerError, StatusUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotAuth, gotQuery string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/users/acema" {
					t.Errorf("path = %q, want %q", r.URL.Path, "/users/acema")
				}
				gotAuth = r.Header.Get("Authorization")
				gotQuery = r.URL.RawQuery
				w.WriteHeader(tt.code)
			}))
			defer srv.Close()

			d := NewDiscogs("test-token")
			d.BaseURL = srv.URL
			d.Client = srv.Client()

			got, err := d.Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if gotAuth != "Discogs token=test-token" {
				t.Errorf("Authorization = %q, want %q", gotAuth, "Discogs token=test-token")
			}
			if strings.Contains(gotQuery, "test-token") {
				t.Errorf("token leaked into the query: %q", gotQuery)
			}
		})
	}
}

// TestTwitchCheck pins the users response mapping, the credentials path (token
// request body, Helix headers), and the login query.
func TestTwitchCheck(t *testing.T) {
	tests := []struct {
		name string
		body string
		want Status
	}{
		{"taken", `{"data":[{"id":"1","login":"acema"}]}`, StatusTaken},
		{"available empty data", `{"data":[]}`, StatusAvailable},
		{"invalid body", `not json`, StatusUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := twitchServer(t, func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/oauth2/token":
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"access_token":"tok","expires_in":3600}`))
				case "/helix/users":
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(tt.body))
				default:
					t.Errorf("unexpected path %q", r.URL.Path)
				}
			})
			defer srv.Close()

			tw := twitchProvider(srv)
			got, err := tw.Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
		})
	}
}

// TestTwitchCheckRequestShape pins the token grant and Helix request details:
// the secret stays in the POST body, while the id and bearer token travel as
// headers.
func TestTwitchCheckRequestShape(t *testing.T) {
	srv := twitchServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			if r.Method != http.MethodPost {
				t.Errorf("token method = %q, want POST", r.Method)
			}
			if err := r.ParseForm(); err != nil {
				t.Fatalf("parse token form: %v", err)
			}
			if got := r.Form.Get("grant_type"); got != "client_credentials" {
				t.Errorf("grant_type = %q, want client_credentials", got)
			}
			if got := r.Form.Get("client_id"); got != "test-id" {
				t.Errorf("client_id = %q, want test-id", got)
			}
			if got := r.Form.Get("client_secret"); got != "test-secret" {
				t.Errorf("client_secret = %q, want test-secret", got)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"access_token":"tok","expires_in":3600}`))
		case "/helix/users":
			if got := r.URL.Query().Get("login"); got != "acema" {
				t.Errorf("login = %q, want acema", got)
			}
			if got := r.Header.Get("Authorization"); got != "Bearer tok" {
				t.Errorf("Authorization = %q, want %q", got, "Bearer tok")
			}
			if got := r.Header.Get("Client-Id"); got != "test-id" {
				t.Errorf("Client-Id = %q, want test-id", got)
			}
			if !strings.Contains(r.URL.RawQuery, "acema") || strings.Contains(r.URL.RawQuery, "secret") {
				t.Errorf("unexpected query %q", r.URL.RawQuery)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":[{"id":"1"}]}`))
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
		}
	})
	defer srv.Close()

	tw := twitchProvider(srv)
	if _, err := tw.Check(context.Background(), "acema"); err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
}

// TestTwitchTokenCache pins that the app token is fetched once and reused
// across checks.
func TestTwitchTokenCache(t *testing.T) {
	var tokenCalls, userCalls int
	srv := twitchServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			tokenCalls++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"access_token":"tok","expires_in":3600}`))
		case "/helix/users":
			userCalls++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":[{"id":"1"}]}`))
		}
	})
	defer srv.Close()

	tw := twitchProvider(srv)
	for i := 0; i < 2; i++ {
		if _, err := tw.Check(context.Background(), "acema"); err != nil {
			t.Fatalf("Check returned error: %v", err)
		}
	}
	if tokenCalls != 1 {
		t.Errorf("token requests = %d, want 1", tokenCalls)
	}
	if userCalls != 2 {
		t.Errorf("user requests = %d, want 2", userCalls)
	}
}

// TestTwitchTokenRefreshOn401 pins that a 401 refreshes the token once and
// retries the lookup.
func TestTwitchTokenRefreshOn401(t *testing.T) {
	var tokenCalls, userCalls int
	srv := twitchServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			tokenCalls++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"access_token":"tok","expires_in":3600}`))
		case "/helix/users":
			userCalls++
			if userCalls == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":[{"id":"1"}]}`))
		}
	})
	defer srv.Close()

	tw := twitchProvider(srv)
	got, err := tw.Check(context.Background(), "acema")
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if got.Status != StatusTaken {
		t.Errorf("status = %q, want %q", got.Status, StatusTaken)
	}
	if tokenCalls != 2 {
		t.Errorf("token requests = %d, want 2", tokenCalls)
	}
	if userCalls != 2 {
		t.Errorf("user requests = %d, want 2", userCalls)
	}
}

// TestTwitchTokenRequestError pins that a failed token grant is surfaced as an
// error rather than a verdict.
func TestTwitchTokenRequestError(t *testing.T) {
	srv := twitchServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	defer srv.Close()

	tw := twitchProvider(srv)
	if _, err := tw.Check(context.Background(), "acema"); err == nil {
		t.Fatal("Check returned nil error, want a token error")
	}
}

// twitchServer serves the token and Helix endpoints on one origin, recording
// requests with handler.
func twitchServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}

// twitchProvider points a provider at a test server's token and Helix paths.
func twitchProvider(srv *httptest.Server) *Twitch {
	tw := NewTwitch("test-id", "test-secret")
	tw.BaseURL = srv.URL + "/helix"
	tw.TokenURL = srv.URL + "/oauth2/token"
	tw.Client = srv.Client()
	return tw
}

// TestCompanyQuery pins the label -> search-term mapping and the exact-name
// comparison used by both company registries.
func TestCompanyQuery(t *testing.T) {
	cases := []struct {
		label string
		query string
	}{
		{"acme", "acme"},
		{"acme-corp", "acme corp"},
		{"acme_corp", "acme corp"},
		{"acme--corp", "acme corp"},
		{"Acme  Corp", "Acme Corp"},
	}
	for _, tt := range cases {
		if got := companyQuery(tt.label); got != tt.query {
			t.Errorf("companyQuery(%q) = %q, want %q", tt.label, got, tt.query)
		}
	}

	matches := []struct {
		label, name string
		want        bool
	}{
		{"acme-corp", "ACME CORP", true},
		{"acme-corp", "Acme  Corp", true},
		{"acme corp", "ACME Corp", true},
		{"acme-corp", "Acme Corporation", false},
		{"acme", "Acme Corp", false},
	}
	for _, tt := range matches {
		if got := companyNameMatches(tt.label, tt.name); got != tt.want {
			t.Errorf("companyNameMatches(%q, %q) = %v, want %v", tt.label, tt.name, got, tt.want)
		}
	}
}

// TestCompaniesHouseCheck pins the exact-name signal (a fuzzy hit stays free)
// and that the API key travels in an HTTP Basic Authorization header, never in
// the URL.
func TestCompaniesHouseCheck(t *testing.T) {
	tests := []struct {
		name string
		code int
		body string
		want Status
	}{
		{"exact title", http.StatusOK, `{"items":[{"title":"Acme Corp"}]}`, StatusTaken},
		{"exact company_name", http.StatusOK, `{"items":[{"company_name":"Acme Corp"}]}`, StatusTaken},
		{"fuzzy only", http.StatusOK, `{"items":[{"title":"Acme Corporation"}]}`, StatusAvailable},
		{"no items", http.StatusOK, `{"items":[]}`, StatusAvailable},
		{"unauthorized", http.StatusUnauthorized, ``, StatusUnknown},
		{"invalid body", http.StatusOK, `not json`, StatusUnknown},
		{"server error", http.StatusInternalServerError, ``, StatusUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotAuth, gotQuery string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/search/companies" {
					t.Errorf("path = %q, want %q", r.URL.Path, "/search/companies")
				}
				gotAuth = r.Header.Get("Authorization")
				gotQuery = r.URL.RawQuery
				w.WriteHeader(tt.code)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			c := NewCompaniesHouse("test-key")
			c.BaseURL = srv.URL
			c.Client = srv.Client()

			got, err := c.Check(context.Background(), "acme-corp")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("test-key:"))
			if gotAuth != wantAuth {
				t.Errorf("Authorization = %q, want %q", gotAuth, wantAuth)
			}
			if got := queryValue(gotQuery, "q"); got != "acme corp" {
				t.Errorf("q = %q, want %q", got, "acme corp")
			}
			if strings.Contains(gotQuery, "test-key") {
				t.Errorf("key leaked into the query: %q", gotQuery)
			}
		})
	}
}

// TestOpenCorporatesCheck pins the exact-name signal and that the token travels
// in the X-API-TOKEN header, never in the URL.
func TestOpenCorporatesCheck(t *testing.T) {
	tests := []struct {
		name string
		code int
		body string
		want Status
	}{
		{"exact", http.StatusOK, `{"results":{"companies":[{"company":{"name":"Acme Corp"}}]}}`, StatusTaken},
		{"fuzzy only", http.StatusOK, `{"results":{"companies":[{"company":{"name":"Acme Corporation"}}]}}`, StatusAvailable},
		{"no companies", http.StatusOK, `{"results":{"companies":[]}}`, StatusAvailable},
		{"unauthorized", http.StatusUnauthorized, ``, StatusUnknown},
		{"invalid body", http.StatusOK, `not json`, StatusUnknown},
		{"server error", http.StatusInternalServerError, ``, StatusUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotToken, gotQuery string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/companies/search" {
					t.Errorf("path = %q, want %q", r.URL.Path, "/companies/search")
				}
				gotToken = r.Header.Get("X-API-TOKEN")
				gotQuery = r.URL.RawQuery
				w.WriteHeader(tt.code)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			o := NewOpenCorporates("test-token")
			o.BaseURL = srv.URL
			o.Client = srv.Client()

			got, err := o.Check(context.Background(), "acme-corp")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if gotToken != "test-token" {
				t.Errorf("X-API-TOKEN = %q, want %q", gotToken, "test-token")
			}
			if got := queryValue(gotQuery, "q"); got != "acme corp" {
				t.Errorf("q = %q, want %q", got, "acme corp")
			}
			if strings.Contains(gotQuery, "test-token") {
				t.Errorf("token leaked into the query: %q", gotQuery)
			}
		})
	}
}

// queryValue parses a raw query and returns the first value for key.
func queryValue(rawQuery, key string) string {
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return ""
	}
	return values.Get(key)
}

// TestDefaultKeyedSources pins that each keyed source is registered only when
// its credentials are present, and that KeyedNames reports them by name.
func TestDefaultKeyedSources(t *testing.T) {
	all := []string{"discogs", "twitch", "companieshouse", "opencorporates"}
	for _, id := range all {
		if _, ok := Default(Config{}).Get(id); ok {
			t.Errorf("Default(Config{}) registered %q with no credentials", id)
		}
	}
	if got := KeyedNames(Config{}); len(got) != 0 {
		t.Errorf("KeyedNames(Config{}) = %v, want none", got)
	}

	discogs := Config{DiscogsToken: "d"}
	if _, ok := Default(discogs).Get("discogs"); !ok {
		t.Error("Default did not register discogs with a token")
	}

	// Twitch needs both halves of the client credentials.
	if _, ok := Default(Config{TwitchClientID: "i"}).Get("twitch"); ok {
		t.Error("Default registered twitch with only a client id")
	}
	if _, ok := Default(Config{TwitchClientSecret: "s"}).Get("twitch"); ok {
		t.Error("Default registered twitch with only a client secret")
	}
	twitch := Config{TwitchClientID: "i", TwitchClientSecret: "s"}
	if _, ok := Default(twitch).Get("twitch"); !ok {
		t.Error("Default did not register twitch with both credentials")
	}

	companiesHouse := Config{CompaniesHouseKey: "k"}
	if _, ok := Default(companiesHouse).Get("companieshouse"); !ok {
		t.Error("Default did not register companieshouse with a key")
	}
	if _, ok := Default(companiesHouse).Get("opencorporates"); ok {
		t.Error("Default registered opencorporates without a token")
	}

	openCorporates := Config{OpenCorporatesToken: "t"}
	if _, ok := Default(openCorporates).Get("opencorporates"); !ok {
		t.Error("Default did not register opencorporates with a token")
	}
	if _, ok := Default(openCorporates).Get("companieshouse"); ok {
		t.Error("Default registered companieshouse without a key")
	}

	cfg := Config{
		DiscogsToken:        "d",
		TwitchClientID:      "i",
		TwitchClientSecret:  "s",
		CompaniesHouseKey:   "k",
		OpenCorporatesToken: "t",
	}
	if got := KeyedNames(cfg); strings.Join(got, ",") != "discogs,twitch,companieshouse,opencorporates" {
		t.Errorf("KeyedNames = %v, want [discogs twitch companieshouse opencorporates]", got)
	}
}

// TestKeyedProvidersUseTimeoutClient asserts each enabled keyed source carries
// a client with a request timeout.
func TestKeyedProvidersUseTimeoutClient(t *testing.T) {
	cfg := Config{
		DiscogsToken:        "d",
		TwitchClientID:      "i",
		TwitchClientSecret:  "s",
		CompaniesHouseKey:   "k",
		OpenCorporatesToken: "t",
	}
	for _, id := range []string{"discogs", "twitch", "companieshouse", "opencorporates"} {
		p, ok := Default(cfg).Get(id)
		if !ok {
			t.Fatalf("keyed source %q is not registered", id)
		}
		client, known := defaultClientOf(p)
		if !known {
			t.Errorf("provider %q has an unknown type; add it to defaultClientOf", id)
			continue
		}
		if client == nil || client.Timeout <= 0 {
			t.Errorf("provider %q has no client timeout", id)
		}
	}
}
