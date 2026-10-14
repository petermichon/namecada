package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
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

// TestXCheck pins the body signal (data present = taken, missing = free) and
// that the bearer token travels in the Authorization header.
func TestXCheck(t *testing.T) {
	tests := []struct {
		name string
		code int
		body string
		want Status
	}{
		{"taken", http.StatusOK, `{"data":{"id":"1","username":"acema"}}`, StatusTaken},
		{"free with errors body", http.StatusOK, `{"errors":[{"title":"Not Found Error"}]}`, StatusAvailable},
		{"free with empty data", http.StatusOK, `{"data":{}}`, StatusAvailable},
		{"available 404", http.StatusNotFound, `{"errors":[{"title":"Not Found Error"}]}`, StatusAvailable},
		{"invalid body", http.StatusOK, `not json`, StatusUnknown},
		{"server error", http.StatusInternalServerError, ``, StatusUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotAuth string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/2/users/by/username/acema" {
					t.Errorf("path = %q, want %q", r.URL.Path, "/2/users/by/username/acema")
				}
				gotAuth = r.Header.Get("Authorization")
				w.WriteHeader(tt.code)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			x := NewX("test-token")
			x.BaseURL = srv.URL
			x.Client = srv.Client()

			got, err := x.Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if gotAuth != "Bearer test-token" {
				t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer test-token")
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

// TestDefaultKeyedSources pins that each keyed source is registered only when
// its credentials are present, and that KeyedNames reports them by name.
func TestDefaultKeyedSources(t *testing.T) {
	all := []string{"discogs", "x", "twitch"}
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
	if _, ok := Default(discogs).Get("x"); ok {
		t.Error("Default registered x without a bearer token")
	}

	x := Config{XBearerToken: "b"}
	if _, ok := Default(x).Get("x"); !ok {
		t.Error("Default did not register x with a bearer token")
	}
	if _, ok := Default(x).Get("discogs"); ok {
		t.Error("Default registered discogs without a token")
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

	cfg := Config{
		DiscogsToken:       "d",
		XBearerToken:       "b",
		TwitchClientID:     "i",
		TwitchClientSecret: "s",
	}
	if got := KeyedNames(cfg); strings.Join(got, ",") != "discogs,x,twitch" {
		t.Errorf("KeyedNames = %v, want [discogs x twitch]", got)
	}
}

// TestKeyedProvidersUseTimeoutClient asserts each enabled keyed source carries
// a client with a request timeout.
func TestKeyedProvidersUseTimeoutClient(t *testing.T) {
	cfg := Config{
		DiscogsToken:       "d",
		XBearerToken:       "b",
		TwitchClientID:     "i",
		TwitchClientSecret: "s",
	}
	for _, id := range []string{"discogs", "x", "twitch"} {
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
