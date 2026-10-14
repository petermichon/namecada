package provider

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
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
				return &Lemmy{BaseURL: baseURL, Client: client}
			},
		},
		{
			name:     "peertube",
			id:       "peertube",
			platform: "peertube",
			wantPath: "/api/v1/accounts/acema",
			build: func(baseURL string, client *http.Client) Provider {
				return &PeerTube{BaseURL: baseURL, Client: client}
			},
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
	if r.Len() != 2 {
		t.Fatalf("Len = %d, want 2", r.Len())
	}
	for _, id := range []string{"lemmy", "peertube"} {
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

func TestMinLength(t *testing.T) {
	tests := []struct {
		name string
		p    Provider
		want int
	}{
		{"known table by id", stubProvider{id: "github"}, 2},
		{"known table by id with a dash", stubProvider{id: "hex-user"}, 3},
		{"unknown id has no minimum", stubProvider{id: "lemmy"}, 0},
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
}

func TestKnownMinLengths(t *testing.T) {
	want := map[string]int{
		"github":      2,
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
