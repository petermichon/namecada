package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRobloxCheck(t *testing.T) {
	tests := []struct {
		name string
		code int
		body string
		want Status
	}{
		{"taken", http.StatusOK, `{"data":[{"requestedUsername":"Roblox","id":1,"name":"Roblox"}]}`, StatusTaken},
		{"free", http.StatusOK, `{"data":[]}`, StatusAvailable},
		{"missing data", http.StatusOK, `{}`, StatusUnknown},
		{"server error", http.StatusInternalServerError, `{"data":[]}`, StatusUnknown},
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

			r := NewRoblox()
			r.BaseURL = srv.URL
			r.Client = srv.Client()

			got, err := r.Check(context.Background(), "Roblox")
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
				Usernames          []string `json:"usernames"`
				ExcludeBannedUsers bool     `json:"excludeBannedUsers"`
			}
			if err := json.Unmarshal(gotBody, &req); err != nil {
				t.Fatalf("request body is not JSON: %v", err)
			}
			if len(req.Usernames) != 1 || req.Usernames[0] != "Roblox" {
				t.Errorf("usernames = %v, want [Roblox]", req.Usernames)
			}
			if req.ExcludeBannedUsers {
				t.Error("excludeBannedUsers = true, want false")
			}
		})
	}
}

func TestRobloxCheckTransportError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close() // no listener: connection refused

	r := NewRoblox()
	r.BaseURL = url
	r.Client = srv.Client()
	if _, err := r.Check(context.Background(), "Roblox"); err == nil {
		t.Fatal("expected a transport error, got nil")
	}
}

func TestRobloxMetadata(t *testing.T) {
	r := NewRoblox()
	if r.ID() != "roblox" {
		t.Errorf("ID = %q, want %q", r.ID(), "roblox")
	}
	if r.Platform() != "roblox" {
		t.Errorf("Platform = %q, want %q", r.Platform(), "roblox")
	}
	if got := r.Name("Roblox"); got != "Roblox" {
		t.Errorf("Name = %q, want %q", got, "Roblox")
	}
	if got := r.URL("Roblox"); got != "https://www.roblox.com/users?username=Roblox" {
		t.Errorf("URL = %q, want %q", got, "https://www.roblox.com/users?username=Roblox")
	}
	if Paid(r) {
		t.Error("Paid = true, want false")
	}
}
