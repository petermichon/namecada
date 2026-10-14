package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestENSCheck(t *testing.T) {
	tests := []struct {
		name string
		code int
		body string
		want Status
	}{
		{"registered", http.StatusOK, `{"address":"0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"}`, StatusTaken},
		{"free", http.StatusOK, `{"address":null}`, StatusAvailable},
		{"empty address", http.StatusOK, `{"address":""}`, StatusAvailable},
		{"server error", http.StatusInternalServerError, `{"address":null}`, StatusUnknown},
		{"invalid json", http.StatusOK, `not json`, StatusUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				w.WriteHeader(tt.code)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			e := NewENS()
			e.BaseURL = srv.URL
			e.Client = srv.Client()

			got, err := e.Check(context.Background(), "vitalik")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if gotPath != "/ens/resolve/vitalik.eth" {
				t.Errorf("path = %q, want %q", gotPath, "/ens/resolve/vitalik.eth")
			}
		})
	}
}

func TestENSCheckTransportError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close() // no listener: connection refused

	e := NewENS()
	e.BaseURL = url
	e.Client = srv.Client()
	if _, err := e.Check(context.Background(), "vitalik"); err == nil {
		t.Fatal("expected a transport error, got nil")
	}
}

func TestENSMetadata(t *testing.T) {
	e := NewENS()
	if e.ID() != "ens" {
		t.Errorf("ID = %q, want %q", e.ID(), "ens")
	}
	if e.Platform() != "ens" {
		t.Errorf("Platform = %q, want %q", e.Platform(), "ens")
	}
	if got := e.Name("vitalik"); got != "vitalik.eth" {
		t.Errorf("Name = %q, want %q", got, "vitalik.eth")
	}
	if got := e.URL("vitalik"); got != "https://app.ens.domains/vitalik.eth" {
		t.Errorf("URL = %q, want %q", got, "https://app.ens.domains/vitalik.eth")
	}
	if !Paid(e) {
		t.Error("Paid = false, want true")
	}
}
