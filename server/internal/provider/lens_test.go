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

func TestLensCheck(t *testing.T) {
	tests := []struct {
		name string
		code int
		body string
		want Status
	}{
		{"taken", http.StatusOK, `{"data":{"account":{"username":{"localName":"stani","value":"lens/stani"}}}}`, StatusTaken},
		{"free", http.StatusOK, `{"data":{"account":null}}`, StatusAvailable},
		{"missing data", http.StatusOK, `{"errors":[{"message":"boom"}]}`, StatusUnknown},
		{"missing account", http.StatusOK, `{"data":{}}`, StatusUnknown},
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

			l := NewLens()
			l.BaseURL = srv.URL
			l.Client = srv.Client()

			got, err := l.Check(context.Background(), "stani")
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
			if req.Variables["name"] != "stani" {
				t.Errorf("variables.name = %q, want %q", req.Variables["name"], "stani")
			}
			if !strings.Contains(req.Query, "account(request: {username: {localName: $name}})") {
				t.Errorf("query = %q, want an account(username.localName:) lookup", req.Query)
			}
		})
	}
}

func TestLensCheckTransportError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close() // no listener: connection refused

	l := NewLens()
	l.BaseURL = url
	l.Client = srv.Client()
	if _, err := l.Check(context.Background(), "stani"); err == nil {
		t.Fatal("expected a transport error, got nil")
	}
}

func TestLensMetadata(t *testing.T) {
	l := NewLens()
	if l.ID() != "lens" {
		t.Errorf("ID = %q, want %q", l.ID(), "lens")
	}
	if l.Platform() != "lens" {
		t.Errorf("Platform = %q, want %q", l.Platform(), "lens")
	}
	if got := l.Name("stani"); got != "stani" {
		t.Errorf("Name = %q, want %q", got, "stani")
	}
	if got := l.URL("stani"); got != "https://hey.xyz/u/stani" {
		t.Errorf("URL = %q, want %q", got, "https://hey.xyz/u/stani")
	}
	if Paid(l) {
		t.Error("Paid = true, want false")
	}
	if got := PlatformName("lens"); got != "Lens" {
		t.Errorf("PlatformName(lens) = %q, want %q", got, "Lens")
	}
}
