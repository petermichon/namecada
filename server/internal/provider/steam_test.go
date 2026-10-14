package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSteamCheck(t *testing.T) {
	tests := []struct {
		name string
		code int
		body string
		want Status
	}{
		{"taken", http.StatusOK, `<?xml version="1.0"?><profile><steamID64>76561197968052866</steamID64></profile>`, StatusTaken},
		{"available", http.StatusOK, `<?xml version="1.0"?><response><error><![CDATA[The specified profile could not be found.]]></error></response>`, StatusAvailable},
		{"unrecognised body", http.StatusOK, `<html>something else</html>`, StatusUnknown},
		{"server error", http.StatusInternalServerError, ``, StatusUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/id/acema" {
					t.Errorf("path = %q, want %q", r.URL.Path, "/id/acema")
				}
				if got := r.URL.Query().Get("xml"); got != "1" {
					t.Errorf("xml = %q, want %q", got, "1")
				}
				w.WriteHeader(tt.code)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			got, err := (&Steam{BaseURL: srv.URL, Client: srv.Client()}).
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

func TestSteamCheckTransportError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close() // no listener: connection refused

	_, err := (&Steam{BaseURL: url, Client: srv.Client()}).
		Check(context.Background(), "acema")
	if err == nil {
		t.Fatal("expected a transport error, got nil")
	}
}

func TestSteamMetadata(t *testing.T) {
	s := NewSteam()
	if s.ID() != "steam" {
		t.Errorf("ID = %q, want %q", s.ID(), "steam")
	}
	if s.Platform() != "steam" {
		t.Errorf("Platform = %q, want %q", s.Platform(), "steam")
	}
	if got := s.Name("acema"); got != "acema" {
		t.Errorf("Name = %q, want %q", got, "acema")
	}
	if got := s.URL("acema"); got != "https://steamcommunity.com/id/acema" {
		t.Errorf("URL = %q, want %q", got, "https://steamcommunity.com/id/acema")
	}
	if Paid(s) {
		t.Error("Paid = true, want false")
	}
}
