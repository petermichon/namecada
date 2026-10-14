package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAniListCheck(t *testing.T) {
	tests := []struct {
		name string
		code int
		body string
		want Status
	}{
		{"taken", http.StatusOK, `{"data":{"User":{"id":6933956}}}`, StatusTaken},
		{"free", http.StatusNotFound, `{"errors":[{"message":"Not Found."}],"data":{"User":null}}`, StatusAvailable},
		{"null user on 200", http.StatusOK, `{"data":{"User":null}}`, StatusAvailable},
		{"missing data", http.StatusOK, `{"errors":[{"message":"boom"}]}`, StatusUnknown},
		{"missing user", http.StatusOK, `{"data":{}}`, StatusUnknown},
		{"server error", http.StatusInternalServerError, `not json`, StatusUnknown},
		{"invalid json", http.StatusOK, `not json`, StatusUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod, gotType string
			var gotBody []byte
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotType = r.Header.Get("Content-Type")
				gotBody, _ = io.ReadAll(r.Body)
				w.WriteHeader(tt.code)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			a := NewAniList()
			a.BaseURL = srv.URL
			a.Client = srv.Client()

			got, err := a.Check(context.Background(), "AniList")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if gotMethod != http.MethodPost {
				t.Errorf("method = %q, want POST", gotMethod)
			}
			if gotType != "application/json" {
				t.Errorf("content-type = %q, want application/json", gotType)
			}
			var req struct {
				Query     string            `json:"query"`
				Variables map[string]string `json:"variables"`
			}
			if err := json.Unmarshal(gotBody, &req); err != nil {
				t.Fatalf("request body is not JSON: %v", err)
			}
			if req.Variables["name"] != "AniList" {
				t.Errorf("variables.name = %q, want %q", req.Variables["name"], "AniList")
			}
			if !strings.Contains(req.Query, "User(name: $name)") {
				t.Errorf("query = %q, want a User(name:) lookup", req.Query)
			}
		})
	}
}

func TestAniListCheckTransportError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close() // no listener: connection refused

	a := NewAniList()
	a.BaseURL = url
	a.Client = srv.Client()
	if _, err := a.Check(context.Background(), "AniList"); err == nil {
		t.Fatal("expected a transport error, got nil")
	}
}

func TestAniListMetadata(t *testing.T) {
	a := NewAniList()
	if a.ID() != "anilist" {
		t.Errorf("ID = %q, want %q", a.ID(), "anilist")
	}
	if a.Platform() != "anilist" {
		t.Errorf("Platform = %q, want %q", a.Platform(), "anilist")
	}
	if got := a.Name("acema"); got != "acema" {
		t.Errorf("Name = %q, want %q", got, "acema")
	}
	if got := a.URL("acema"); got != "https://anilist.co/user/acema" {
		t.Errorf("URL = %q, want %q", got, "https://anilist.co/user/acema")
	}
	if Paid(a) {
		t.Error("Paid = true, want false")
	}
}
