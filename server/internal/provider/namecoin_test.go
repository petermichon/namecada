package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNamecoinCheck(t *testing.T) {
	tests := []struct {
		name string
		code int
		body string
		want Status
	}{
		{"taken", http.StatusOK, `{"name":"d/acema","expired":false}`, StatusTaken},
		{"expired", http.StatusOK, `{"name":"d/acema","expired":true}`, StatusAvailable},
		{"never registered", http.StatusNotFound, ``, StatusAvailable},
		{"missing expired", http.StatusOK, `{"name":"d/acema"}`, StatusUnknown},
		{"null expired", http.StatusOK, `{"name":"d/acema","expired":null}`, StatusUnknown},
		{"invalid json", http.StatusOK, `not json`, StatusUnknown},
		{"server error", http.StatusInternalServerError, ``, StatusUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.EscapedPath()
				w.WriteHeader(tt.code)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			n := NewNamecoin()
			n.BaseURL = srv.URL
			n.Client = srv.Client()

			got, err := n.Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if gotPath != "/api/name/d%2Facema" {
				t.Errorf("path = %q, want %q", gotPath, "/api/name/d%2Facema")
			}
		})
	}
}

func TestNamecoinCheckTransportError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close() // no listener: connection refused

	n := NewNamecoin()
	n.BaseURL = url
	n.Client = srv.Client()
	if _, err := n.Check(context.Background(), "acema"); err == nil {
		t.Fatal("expected a transport error, got nil")
	}
}

func TestNamecoinMetadata(t *testing.T) {
	n := NewNamecoin()
	if n.ID() != "namecoin" {
		t.Errorf("ID = %q, want %q", n.ID(), "namecoin")
	}
	if n.Platform() != "namecoin" {
		t.Errorf("Platform = %q, want %q", n.Platform(), "namecoin")
	}
	if got := n.Name("acema"); got != "acema" {
		t.Errorf("Name = %q, want %q", got, "acema")
	}
	if got := n.URL("acema"); got != "https://explorer.namecoin.co/name/d/acema" {
		t.Errorf("URL = %q, want %q", got, "https://explorer.namecoin.co/name/d/acema")
	}
	if !Paid(n) {
		t.Error("Paid = false, want true")
	}
	if got := PlatformName(n.Platform()); got != "Namecoin" {
		t.Errorf("PlatformName = %q, want %q", got, "Namecoin")
	}
}
