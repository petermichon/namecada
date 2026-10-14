package provider

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// httpSource builds a provider pointed at a test server.
type httpSource struct {
	name     string
	id       string
	platform string
	wantPath string
	wantQ    string
	build    func(baseURL string, client *http.Client) Provider
}

func httpSources() []httpSource {
	return []httpSource{
		{
			name:     "lemmy",
			id:       "lemmy",
			platform: "lemmy",
			wantPath: "/api/v3/user",
			wantQ:    "username=acema",
			build: func(baseURL string, client *http.Client) Provider {
				p := NewLemmy()
				p.BaseURL = baseURL
				p.Client = client
				return p
			},
		},
		endpointSource("npm", "npm", "/acema", "", NewNPM),
		endpointSource("pypi", "pypi", "/pypi/acema/json", "", NewPyPI),
		endpointSource("crates", "crates", "/api/v1/crates/acema", "", NewCrates),
		endpointSource("nuget", "nuget", "/v3-flatcontainer/acema/index.json", "", NewNuGet),
		endpointSource("rubygems", "rubygems", "/api/v1/gems/acema.json", "", NewRubyGems),
		endpointSource("hex", "hex", "/api/packages/acema", "", NewHex),
		endpointSource("hex-user", "hex", "/api/users/acema", "", NewHexUser),
		endpointSource("pub", "pub", "/api/packages/acema", "", NewPub),
		endpointSource("homebrew", "homebrew", "/api/formula/acema.json", "", NewHomebrew),
		endpointSource("hackage", "hackage", "/package/acema", "", NewHackage),
		endpointSource("clojars", "clojars", "/api/artifacts/acema", "", NewClojars),
		endpointSource("chocolatey", "chocolatey", "/packages/acema", "", NewChocolatey),
		endpointSource("aur", "aur", "/packages/acema", "", NewAUR),
		endpointSource("firefox", "firefox", "/api/v5/addons/addon/acema/", "", NewFirefox),
		endpointSource("jsr", "jsr", "/api/scopes/acema", "", NewJSR),
		endpointSource("openvsx", "openvsx", "/api/acema", "", NewOpenVSX),
		endpointSource("github", "github", "/users/acema", "", NewGitHub),
		endpointSource("bitbucket", "bitbucket", "/2.0/workspaces/acema", "", NewBitbucket),
		{
			name:     "wordpress",
			id:       "wordpress",
			platform: "wordpress",
			wantPath: "/plugins/info/1.2/",
			wantQ:    "action=plugin_information&request%5Bslug%5D=acema",
			build: func(baseURL string, client *http.Client) Provider {
				p := NewWordPress()
				p.BaseURL = baseURL
				p.Client = client
				return p
			},
		},
		endpointSource("farcaster", "farcaster", "/transfers/current", "name=acema", NewFarcaster),
		endpointSource("lichess", "lichess", "/api/user/acema", "", NewLichess),
		endpointSource("chesscom", "chesscom", "/pub/player/acema", "", NewChesscom),
		endpointSource("huggingface", "huggingface", "/api/users/acema/overview", "", NewHuggingFace),
		endpointSource("dailymotion", "dailymotion", "/user/acema", "", NewDailymotion),
		endpointSource("mixcloud", "mixcloud", "/acema/", "", NewMixcloud),
		endpointSource("mastodon", "mastodon", "/@acema", "", NewMastodon),
	}
}

// endpointSource builds a table entry for a status-only registry provider,
// pointing its constructor at a test server.
func endpointSource(id, platform, wantPath, wantQ string, newProvider func() Provider) httpSource {
	return httpSource{
		name:     id,
		id:       id,
		platform: platform,
		wantPath: wantPath,
		wantQ:    wantQ,
		build: func(baseURL string, client *http.Client) Provider {
			e := newProvider().(*endpoint)
			e.baseURL = baseURL
			e.client = client
			return e
		},
	}
}

func TestProvidersCheck(t *testing.T) {
	statuses := []struct {
		name string
		code int
		want Status
	}{
		{"taken", http.StatusOK, StatusTaken},
		{"available", http.StatusNotFound, StatusAvailable},
		{"unknown", http.StatusInternalServerError, StatusUnknown},
	}

	for _, src := range httpSources() {
		for _, st := range statuses {
			t.Run(src.name+"/"+st.name, func(t *testing.T) {
				var gotPath, gotQuery string
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					gotPath = r.URL.Path
					gotQuery = r.URL.RawQuery
					w.WriteHeader(st.code)
				}))
				defer srv.Close()

				got, err := src.build(srv.URL, srv.Client()).Check(context.Background(), "acema")
				if err != nil {
					t.Fatalf("Check returned error: %v", err)
				}
				if got.Status != st.want {
					t.Errorf("status = %q, want %q", got.Status, st.want)
				}
				if gotPath != src.wantPath {
					t.Errorf("path = %q, want %q", gotPath, src.wantPath)
				}
				if gotQuery != src.wantQ {
					t.Errorf("query = %q, want %q", gotQuery, src.wantQ)
				}
			})
		}
	}
}

func TestProvidersUserAgent(t *testing.T) {
	for _, src := range httpSources() {
		t.Run(src.name, func(t *testing.T) {
			var gotUA string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotUA = r.Header.Get("User-Agent")
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			if _, err := src.build(srv.URL, srv.Client()).Check(context.Background(), "acema"); err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if gotUA != UserAgent {
				t.Errorf("User-Agent = %q, want %q", gotUA, UserAgent)
			}
		})
	}
}

func TestProvidersCheckTransportError(t *testing.T) {
	for _, src := range httpSources() {
		t.Run(src.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
			url := srv.URL
			srv.Close() // no listener: connection refused

			_, err := src.build(url, srv.Client()).Check(context.Background(), "acema")
			if err == nil {
				t.Fatal("expected a transport error, got nil")
			}
		})
	}
}

func TestJSRCodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	e := NewJSR().(*endpoint)
	e.baseURL = srv.URL
	e.client = srv.Client()

	got, err := e.Check(context.Background(), "acema")
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if got.Status != StatusTaken {
		t.Errorf("status = %q, want %q", got.Status, StatusTaken)
	}
}

func TestFirefoxCheck(t *testing.T) {
	tests := []struct {
		name string
		code int
		want Status
	}{
		{"taken", http.StatusOK, StatusTaken},
		{"restricted", http.StatusUnauthorized, StatusTaken},
		{"available", http.StatusNotFound, StatusAvailable},
		{"server error", http.StatusInternalServerError, StatusUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v5/addons/addon/acema/" {
					t.Errorf("path = %q, want %q", r.URL.Path, "/api/v5/addons/addon/acema/")
				}
				w.WriteHeader(tt.code)
			}))
			defer srv.Close()

			e := NewFirefox().(*endpoint)
			e.baseURL = srv.URL
			e.client = srv.Client()

			got, err := e.Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
		})
	}
}

func TestWordPressCheck(t *testing.T) {
	tests := []struct {
		name string
		code int
		body string
		want Status
	}{
		{"published", http.StatusOK, `{}`, StatusTaken},
		{"closed", http.StatusNotFound, `{"closed":true}`, StatusTaken},
		{"free", http.StatusNotFound, `{"closed":false}`, StatusAvailable},
		{"free without body", http.StatusNotFound, ``, StatusAvailable},
		{"server error", http.StatusInternalServerError, ``, StatusUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.code)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			got, err := (&WordPress{BaseURL: srv.URL, Client: srv.Client()}).
				Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
		})
	}
}

func TestGitLabCheck(t *testing.T) {
	tests := []struct {
		name      string
		usersCode int
		usersBody string
		groupCode int
		want      Status
		wantGroup bool
	}{
		{"user exists", http.StatusOK, `[{"id":1}]`, http.StatusOK, StatusTaken, false},
		{"group takes empty users", http.StatusOK, `[]`, http.StatusOK, StatusTaken, true},
		{"free name", http.StatusOK, `[]`, http.StatusNotFound, StatusAvailable, true},
		{"users error", http.StatusInternalServerError, ``, http.StatusOK, StatusUnknown, false},
		{"users non-array body", http.StatusOK, `{"message":"boom"}`, http.StatusOK, StatusUnknown, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var groupCalled bool
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/v4/users":
					if got := r.URL.Query().Get("username"); got != "acema" {
						t.Errorf("username = %q, want %q", got, "acema")
					}
					w.WriteHeader(tt.usersCode)
					_, _ = w.Write([]byte(tt.usersBody))
				case "/api/v4/groups/acema":
					groupCalled = true
					w.WriteHeader(tt.groupCode)
				default:
					t.Errorf("unexpected path %q", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer srv.Close()

			got, err := (&GitLab{BaseURL: srv.URL, Client: srv.Client()}).
				Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if groupCalled != tt.wantGroup {
				t.Errorf("group fallback called = %v, want %v", groupCalled, tt.wantGroup)
			}
		})
	}
}

func TestCodebergCheck(t *testing.T) {
	tests := []struct {
		name string
		code int
		body string
		want Status
	}{
		{"taken", http.StatusOK, `{}`, StatusTaken},
		{"available", http.StatusNotFound, `{"message":"user acema does not exist"}`, StatusAvailable},
		{"hidden account", http.StatusNotFound, `{"message":"user is not visible"}`, StatusTaken},
		{"404 without message", http.StatusNotFound, `{}`, StatusUnknown},
		{"404 invalid body", http.StatusNotFound, `not json`, StatusUnknown},
		{"server error", http.StatusInternalServerError, ``, StatusUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/users/acema" {
					t.Errorf("path = %q, want %q", r.URL.Path, "/api/v1/users/acema")
				}
				w.WriteHeader(tt.code)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			got, err := (&Codeberg{BaseURL: srv.URL, Client: srv.Client()}).
				Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
		})
	}
}

func TestWikipediaCheck(t *testing.T) {
	tests := []struct {
		name string
		code int
		body string
		want Status
	}{
		{"taken", http.StatusOK, `{"query":{"globaluserinfo":{"id":1,"name":"acema"}}}`, StatusTaken},
		{"available", http.StatusOK, `{"query":{"globaluserinfo":{"missing":"1"}}}`, StatusAvailable},
		{"no query", http.StatusOK, `{}`, StatusUnknown},
		{"no info", http.StatusOK, `{"query":{}}`, StatusUnknown},
		{"null info", http.StatusOK, `{"query":{"globaluserinfo":null}}`, StatusUnknown},
		{"invalid body", http.StatusOK, `not json`, StatusUnknown},
		{"server error", http.StatusInternalServerError, ``, StatusUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotQuery url.Values
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/w/api.php" {
					t.Errorf("path = %q, want %q", r.URL.Path, "/w/api.php")
				}
				gotQuery = r.URL.Query()
				w.WriteHeader(tt.code)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			got, err := (&Wikipedia{BaseURL: srv.URL, Client: srv.Client()}).
				Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if got := gotQuery.Get("guiuser"); got != "acema" {
				t.Errorf("guiuser = %q, want %q", got, "acema")
			}
			if got := gotQuery.Get("meta"); got != "globaluserinfo" {
				t.Errorf("meta = %q, want %q", got, "globaluserinfo")
			}
		})
	}
}

func TestDevToCheck(t *testing.T) {
	tests := []struct {
		name     string
		userCode int
		orgCode  int
		want     Status
		wantOrg  bool
	}{
		{"user taken", http.StatusOK, http.StatusOK, StatusTaken, false},
		{"org taken", http.StatusNotFound, http.StatusOK, StatusTaken, true},
		{"free name", http.StatusNotFound, http.StatusNotFound, StatusAvailable, true},
		{"user unknown", http.StatusInternalServerError, http.StatusOK, StatusUnknown, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var orgCalled bool
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/users/by_username":
					if got := r.URL.Query().Get("url"); got != "acema" {
						t.Errorf("url = %q, want %q", got, "acema")
					}
					w.WriteHeader(tt.userCode)
				case "/api/organizations/acema":
					orgCalled = true
					w.WriteHeader(tt.orgCode)
				default:
					t.Errorf("unexpected path %q", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer srv.Close()

			got, err := (&DevTo{BaseURL: srv.URL, Client: srv.Client()}).
				Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if orgCalled != tt.wantOrg {
				t.Errorf("org fallback called = %v, want %v", orgCalled, tt.wantOrg)
			}
		})
	}
}

func TestMinehutCheck(t *testing.T) {
	tests := []struct {
		name string
		code int
		body string
		want Status
	}{
		{"taken", http.StatusOK, `{"server":{"name":"acema"}}`, StatusTaken},
		{"available", http.StatusOK, `{"server":null}`, StatusAvailable},
		{"absent field", http.StatusOK, `{}`, StatusUnknown},
		{"invalid body", http.StatusOK, `not json`, StatusUnknown},
		{"server error", http.StatusInternalServerError, ``, StatusUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/server/acema" {
					t.Errorf("path = %q, want %q", r.URL.Path, "/server/acema")
				}
				if got := r.URL.Query().Get("byName"); got != "true" {
					t.Errorf("byName = %q, want %q", got, "true")
				}
				w.WriteHeader(tt.code)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			got, err := (&Minehut{BaseURL: srv.URL, Client: srv.Client()}).
				Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
		})
	}
}

func TestBlueskyCheck(t *testing.T) {
	tests := []struct {
		name string
		code int
		body string
		want Status
	}{
		{"available", http.StatusOK, `{"result":{"$type":"com.atproto.temp.defs#resultAvailable"}}`, StatusAvailable},
		{"taken", http.StatusOK, `{"result":{"$type":"com.atproto.temp.defs#resultUnavailable"}}`, StatusTaken},
		{"missing type", http.StatusOK, `{"result":{}}`, StatusUnknown},
		{"null body", http.StatusOK, `null`, StatusUnknown},
		{"invalid body", http.StatusOK, `not json`, StatusUnknown},
		{"server error", http.StatusInternalServerError, ``, StatusUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/xrpc/com.atproto.temp.checkHandleAvailability" {
					t.Errorf("path = %q", r.URL.Path)
				}
				if got := r.URL.Query().Get("handle"); got != "acema.bsky.social" {
					t.Errorf("handle = %q, want %q", got, "acema.bsky.social")
				}
				w.WriteHeader(tt.code)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			got, err := (&Bluesky{BaseURL: srv.URL, Client: srv.Client()}).
				Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
		})
	}
}

func TestOpenCollectiveCheck(t *testing.T) {
	tests := []struct {
		name     string
		codes    []int
		want     Status
		wantCall int
	}{
		{"taken", []int{http.StatusOK}, StatusTaken, 1},
		{"available", []int{http.StatusNotFound}, StatusAvailable, 1},
		{"retry taken", []int{http.StatusInternalServerError, http.StatusOK}, StatusTaken, 2},
		{"retry available", []int{http.StatusInternalServerError, http.StatusNotFound}, StatusAvailable, 2},
		{"retry unknown", []int{http.StatusInternalServerError, http.StatusInternalServerError}, StatusUnknown, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls int
			var queries []url.Values
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				queries = append(queries, r.URL.Query())
				code := tt.codes[calls]
				calls++
				w.WriteHeader(code)
			}))
			defer srv.Close()

			oc := &OpenCollective{
				BaseURL: srv.URL,
				Client:  srv.Client(),
				Now:     func() time.Time { return time.UnixMilli(1700000000000) },
			}
			got, err := oc.Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if calls != tt.wantCall {
				t.Fatalf("calls = %d, want %d", calls, tt.wantCall)
			}
			if _, ok := queries[0]["_"]; ok {
				t.Error("first request unexpectedly had a cache-buster")
			}
			if tt.wantCall == 2 {
				if got := queries[1].Get("_"); got != "1700000000000" {
					t.Errorf("cache-buster = %q, want %q", got, "1700000000000")
				}
			}
		})
	}
}

func TestMastodonCheck(t *testing.T) {
	tests := []struct {
		code int
		want Status
	}{
		{http.StatusOK, StatusTaken},
		{http.StatusGone, StatusTaken},
		{http.StatusNotFound, StatusAvailable},
		// The HTML profile route answers 401 to a JSON Accept; Mastodon must
		// not read that as taken the way Firefox's API does.
		{http.StatusUnauthorized, StatusUnknown},
		{http.StatusInternalServerError, StatusUnknown},
	}
	for _, tt := range tests {
		t.Run(http.StatusText(tt.code), func(t *testing.T) {
			var gotAccept string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/@acema" {
					t.Errorf("path = %q, want %q", r.URL.Path, "/@acema")
				}
				gotAccept = r.Header.Get("Accept")
				w.WriteHeader(tt.code)
			}))
			defer srv.Close()

			e := NewMastodon().(*endpoint)
			e.baseURL = srv.URL
			e.client = srv.Client()

			got, err := e.Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if gotAccept != "text/html" {
				t.Errorf("Accept = %q, want %q", gotAccept, "text/html")
			}
		})
	}
}

// TestPeerTubeCheck pins that a name is taken when either the account or the
// video-channel endpoint answers 200, and free only when both answer 404.
func TestPeerTubeCheck(t *testing.T) {
	tests := []struct {
		name        string
		accountCode int
		channelCode int
		want        Status
		wantChannel bool
		wantURL     string
	}{
		{"account taken", http.StatusOK, http.StatusOK, StatusTaken, false, "https://framatube.org/a/acema"},
		{"channel taken", http.StatusNotFound, http.StatusOK, StatusTaken, true, "https://framatube.org/c/acema"},
		{"both free", http.StatusNotFound, http.StatusNotFound, StatusAvailable, true, ""},
		{"account free channel unknown", http.StatusNotFound, http.StatusInternalServerError, StatusUnknown, true, ""},
		{"account unknown channel free", http.StatusInternalServerError, http.StatusNotFound, StatusUnknown, true, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var channelCalled bool
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/v1/accounts/acema":
					w.WriteHeader(tt.accountCode)
				case "/api/v1/video-channels/acema":
					channelCalled = true
					w.WriteHeader(tt.channelCode)
				default:
					t.Errorf("unexpected path %q", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer srv.Close()

			p := NewPeerTube()
			p.BaseURL = srv.URL
			p.Client = srv.Client()

			got, err := p.Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if got.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", got.URL, tt.wantURL)
			}
			if channelCalled != tt.wantChannel {
				t.Errorf("channel endpoint called = %v, want %v", channelCalled, tt.wantChannel)
			}
		})
	}
}

// TestPeerTubeMetadata covers the metadata and User-Agent contract that the
// shared table tests no longer exercise now that PeerTube makes two requests.
func TestPeerTubeMetadata(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p := NewPeerTube()
	p.BaseURL = srv.URL
	p.Client = srv.Client()

	if p.ID() != "peertube" || p.Platform() != "peertube" {
		t.Errorf("metadata = %q/%q, want peertube/peertube", p.ID(), p.Platform())
	}
	if got := p.Name("acema"); got != "acema" {
		t.Errorf("Name = %q, want %q", got, "acema")
	}
	if p.URL("acema") == "" {
		t.Error("URL is empty")
	}
	if _, err := p.Check(context.Background(), "acema"); err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if gotUA != UserAgent {
		t.Errorf("User-Agent = %q, want %q", gotUA, UserAgent)
	}
}

// TestPeerTubeTransportError pins that a failed request is surfaced as an error
// rather than folded into an available/taken verdict.
func TestPeerTubeTransportError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close() // no listener: connection refused

	_, err := (&PeerTube{BaseURL: url, Client: srv.Client()}).
		Check(context.Background(), "acema")
	if err == nil {
		t.Fatal("expected a transport error, got nil")
	}
}

func TestEndpointDefaultAccept(t *testing.T) {
	var gotAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAccept = r.Header.Get("Accept")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	e := NewNPM().(*endpoint)
	e.baseURL = srv.URL
	e.client = srv.Client()

	if _, err := e.Check(context.Background(), "acema"); err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if gotAccept != defaultAccept {
		t.Errorf("Accept = %q, want %q", gotAccept, defaultAccept)
	}
}

func TestProvidersMetadata(t *testing.T) {
	for _, src := range httpSources() {
		t.Run(src.name, func(t *testing.T) {
			p := src.build("https://example.test", http.DefaultClient)
			if p.ID() != src.id {
				t.Errorf("ID = %q, want %q", p.ID(), src.id)
			}
			if p.Platform() != src.platform {
				t.Errorf("Platform = %q, want %q", p.Platform(), src.platform)
			}
			if got := p.Name("acema"); got != "acema" {
				t.Errorf("Name = %q, want %q", got, "acema")
			}
			if got := p.URL("acema"); got == "" {
				t.Error("URL is empty")
			}
			if Paid(p) {
				t.Errorf("Paid = true, want free for registry/handle sources")
			}
		})
	}
}

func TestCheckHTTPContextCanceled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewLemmy().Check(ctx, "acema")
	if err == nil {
		t.Fatal("expected an error from a canceled context, got nil")
	}
}

func TestDefaultClientHasTimeout(t *testing.T) {
	if DefaultTimeout <= 0 {
		t.Fatalf("DefaultTimeout = %v, want a positive timeout", DefaultTimeout)
	}
	if DefaultClient.Timeout != DefaultTimeout {
		t.Errorf("DefaultClient.Timeout = %v, want %v", DefaultClient.Timeout, DefaultTimeout)
	}
}

// defaultClientOf returns the request client a built-in provider holds. The
// second result reports whether the provider's type is known, so the test fails
// when a new provider is added without a timeout-bound client.
func defaultClientOf(p Provider) (*http.Client, bool) {
	switch v := p.(type) {
	case *endpoint:
		return v.client, true
	case *Domain:
		return v.Client, true
	case *Lemmy:
		return v.Client, true
	case *PeerTube:
		return v.Client, true
	case *WordPress:
		return v.Client, true
	case *GitLab:
		return v.Client, true
	case *Codeberg:
		return v.Client, true
	case *Wikipedia:
		return v.Client, true
	case *DevTo:
		return v.Client, true
	case *Minehut:
		return v.Client, true
	case *Bluesky:
		return v.Client, true
	case *OpenCollective:
		return v.Client, true
	}
	return nil, false
}

// TestDefaultProvidersUseTimeoutClient asserts every registered provider is
// bound to a client that carries a request timeout, so one hung source cannot
// stall a scan.
func TestDefaultProvidersUseTimeoutClient(t *testing.T) {
	for _, p := range Default().All() {
		client, known := defaultClientOf(p)
		if !known {
			t.Errorf("provider %q has an unknown type; add it to defaultClientOf", p.ID())
			continue
		}
		if client == nil || client.Timeout <= 0 {
			t.Errorf("provider %q has no client timeout", p.ID())
		}
	}
}

// TestCheckHTTPClientTimeout proves a request timeout aborts a hung source
// rather than waiting for it to answer.
func TestCheckHTTPClientTimeout(t *testing.T) {
	block := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-block:
		case <-r.Context().Done():
		}
	}))
	defer srv.Close()
	defer close(block)

	client := &http.Client{Timeout: 50 * time.Millisecond}
	if _, err := checkHTTP(context.Background(), client, srv.URL, defaultCodes); err == nil {
		t.Fatal("expected a timeout error, got nil")
	}
}

func TestRegistry(t *testing.T) {
	t.Run("register and retrieve in order", func(t *testing.T) {
		r := NewRegistry()
		lemmy := NewLemmy()
		peertube := NewPeerTube()
		if err := r.Register(lemmy); err != nil {
			t.Fatalf("register lemmy: %v", err)
		}
		if err := r.Register(peertube); err != nil {
			t.Fatalf("register peertube: %v", err)
		}
		if r.Len() != 2 {
			t.Fatalf("Len = %d, want 2", r.Len())
		}
		got := r.All()
		if len(got) != 2 || got[0].ID() != "lemmy" || got[1].ID() != "peertube" {
			t.Fatalf("All order = %v, want [lemmy peertube]", ids(got))
		}
		if p, ok := r.Get("peertube"); !ok || p.ID() != "peertube" {
			t.Fatalf("Get(peertube) = %v, %v", p, ok)
		}
		if _, ok := r.Get("missing"); ok {
			t.Fatal("Get(missing) unexpectedly found a provider")
		}
	})

	t.Run("rejects duplicate ids", func(t *testing.T) {
		r := NewRegistry()
		if err := r.Register(NewLemmy()); err != nil {
			t.Fatalf("first register: %v", err)
		}
		err := r.Register(NewLemmy())
		if !errors.Is(err, ErrDuplicate) {
			t.Fatalf("err = %v, want ErrDuplicate", err)
		}
		if r.Len() != 1 {
			t.Fatalf("Len = %d, want 1", r.Len())
		}
	})

	t.Run("rejects nil", func(t *testing.T) {
		if err := NewRegistry().Register(nil); err == nil {
			t.Fatal("expected an error registering nil")
		}
	})
}

func TestDefaultRegistry(t *testing.T) {
	r := Default()
	if r.Len() != 41 {
		t.Fatalf("Len = %d, want 41", r.Len())
	}
	for _, id := range []string{
		"domain.com", "domain.org", "domain.dev", "domain.io", "domain.ai", "domain.app",
		"lemmy", "peertube",
		"npm", "pypi", "crates", "nuget", "rubygems", "hex", "hex-user", "pub",
		"homebrew", "hackage", "clojars", "chocolatey", "aur", "firefox", "jsr", "wordpress", "openvsx",
		"github", "gitlab", "bitbucket", "codeberg",
		"wikipedia", "devto", "farcaster", "minehut", "lichess", "chesscom",
		"huggingface", "dailymotion", "mixcloud", "bluesky", "mastodon",
		"opencollective",
	} {
		if _, ok := r.Get(id); !ok {
			t.Errorf("default registry is missing %q", id)
		}
	}
}

// stubProvider implements only Provider, so it has no explicit minimum.
type stubProvider struct{ id string }

func (s stubProvider) ID() string                                    { return s.id }
func (s stubProvider) Platform() string                              { return s.id }
func (s stubProvider) Name(label string) string                      { return label }
func (s stubProvider) URL(label string) string                       { return label }
func (s stubProvider) Check(context.Context, string) (Result, error) { return Result{}, nil }

// stubWithMin additionally declares its own minimum length.
type stubWithMin struct {
	stubProvider
	min int
}

func (s stubWithMin) MinLength() int { return s.min }

// stubWithMax additionally declares its own maximum length.
type stubWithMax struct {
	stubProvider
	max int
}

func (s stubWithMax) MaxLength() int { return s.max }

// stubWithPaid additionally declares whether registering costs money.
type stubWithPaid struct {
	stubProvider
	paid bool
}

func (s stubWithPaid) Paid() bool { return s.paid }

func TestPaid(t *testing.T) {
	tests := []struct {
		name string
		p    Provider
		want bool
	}{
		{"domain is paid", NewDomain(".com"), true},
		{"free registry is not paid", NewNPM(), false},
		{"nil is not paid", nil, false},
		{"provider without the method is free", stubProvider{id: "custom"}, false},
		{"explicit paid wins", stubWithPaid{stubProvider{id: "custom"}, true}, true},
		{"explicit free stays free", stubWithPaid{stubProvider{id: "domains"}, false}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Paid(tt.p); got != tt.want {
				t.Errorf("Paid = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDefaultPaidSources pins that the domain TLDs are the only paid sources.
func TestDefaultPaidSources(t *testing.T) {
	for _, p := range Default().All() {
		want := p.Platform() == "domains"
		if got := Paid(p); got != want {
			t.Errorf("Paid(%q) = %v, want %v", p.ID(), got, want)
		}
	}
}

func TestMinLength(t *testing.T) {
	tests := []struct {
		name string
		p    Provider
		want int
	}{
		{"known table by id", stubProvider{id: "github"}, 2},
		{"known table by id with a dash", stubProvider{id: "hex-user"}, 3},
		{"lemmy has a minimum", stubProvider{id: "lemmy"}, 3},
		{"unknown id has no minimum", stubProvider{id: "nonexistent"}, 0},
		{"explicit minimum wins", stubWithMin{stubProvider{id: "github"}, 7}, 7},
		{"explicit zero disables the table", stubWithMin{stubProvider{id: "github"}, 0}, 0},
		{"nil has no minimum", nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MinLength(tt.p); got != tt.want {
				t.Errorf("MinLength = %d, want %d", got, tt.want)
			}
		})
	}
	if got := MinLength(stubProvider{id: "lemmy"}); got <= len("ab") {
		t.Errorf("MinLength(lemmy) = %d, want > len(\"ab\")", got)
	}
}

func TestKnownMinLengths(t *testing.T) {
	want := map[string]int{
		"github":      2,
		"lemmy":       3,
		"hex":         2,
		"hex-user":    3,
		"wordpress":   5,
		"bitbucket":   4,
		"minehut":     4,
		"lichess":     2,
		"chesscom":    3,
		"huggingface": 2,
		"dailymotion": 3,
		"bluesky":     3,
		"devto":       2,
	}
	if len(minLengths) != len(want) {
		t.Fatalf("known minimums has %d entries, want %d", len(minLengths), len(want))
	}
	for id, minimum := range want {
		if got := minLengths[id]; got != minimum {
			t.Errorf("minLengths[%q] = %d, want %d", id, got, minimum)
		}
	}
}

func TestMaxLength(t *testing.T) {
	tests := []struct {
		name string
		p    Provider
		want int
	}{
		{"known table by id", stubProvider{id: "lemmy"}, 20},
		{"unknown id has no maximum", stubProvider{id: "nonexistent"}, 0},
		{"explicit maximum wins", stubWithMax{stubProvider{id: "lemmy"}, 12}, 12},
		{"explicit zero disables the table", stubWithMax{stubProvider{id: "lemmy"}, 0}, 0},
		{"nil has no maximum", nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaxLength(tt.p); got != tt.want {
				t.Errorf("MaxLength = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestKnownMaxLengths(t *testing.T) {
	want := map[string]int{
		"lemmy": 20,
	}
	if len(maxLengths) != len(want) {
		t.Fatalf("known maximums has %d entries, want %d", len(maxLengths), len(want))
	}
	for id, maximum := range want {
		if got := maxLengths[id]; got != maximum {
			t.Errorf("maxLengths[%q] = %d, want %d", id, got, maximum)
		}
	}
}

func TestCodeMapClassify(t *testing.T) {
	codes := codeMap{available: []int{404}, taken: []int{200, 401}}
	tests := []struct {
		status int
		want   Status
	}{
		{200, StatusTaken},
		{401, StatusTaken},
		{404, StatusAvailable},
		{500, StatusUnknown},
		{0, StatusUnknown},
	}
	for _, tt := range tests {
		if got := codes.classify(tt.status); got.Status != tt.want {
			t.Errorf("classify(%d) = %q, want %q", tt.status, got.Status, tt.want)
		}
	}
}

func ids(ps []Provider) []string {
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = p.ID()
	}
	return out
}
