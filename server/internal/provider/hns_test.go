package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHNSCheck(t *testing.T) {
	tests := []struct {
		name string
		code int
		body string
		want Status
	}{
		{"closed", http.StatusOK, `{"info":{"state":"CLOSED","registered":true}}`, StatusTaken},
		{"available", http.StatusOK, `{"info":{"state":"AVAILABLE","registered":true}}`, StatusAvailable},
		{"unknown state", http.StatusOK, `{"info":{"state":"UNKNOWN","registered":false}}`, StatusAvailable},
		{"missing info", http.StatusOK, `{}`, StatusUnknown},
		{"null info", http.StatusOK, `{"info":null}`, StatusUnknown},
		{"unrecognised state", http.StatusOK, `{"info":{"state":"WEIRD"}}`, StatusUnknown},
		{"reserved", http.StatusBadGateway, ``, StatusUnknown},
		{"server error", http.StatusInternalServerError, ``, StatusUnknown},
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

			h := NewHNS()
			h.BaseURL = srv.URL
			h.Client = srv.Client()

			got, err := h.Check(context.Background(), "hnsstat")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if gotPath != "/api/names/hnsstat/details" {
				t.Errorf("path = %q, want %q", gotPath, "/api/names/hnsstat/details")
			}
		})
	}
}

func TestHNSCheckTransportError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close() // no listener: connection refused

	h := NewHNS()
	h.BaseURL = url
	h.Client = srv.Client()
	if _, err := h.Check(context.Background(), "hnsstat"); err == nil {
		t.Fatal("expected a transport error, got nil")
	}
}

func TestHNSMetadata(t *testing.T) {
	h := NewHNS()
	if h.ID() != "hns" {
		t.Errorf("ID = %q, want %q", h.ID(), "hns")
	}
	if h.Platform() != "hns" {
		t.Errorf("Platform = %q, want %q", h.Platform(), "hns")
	}
	if got := h.Name("hnsstat"); got != "hnsstat" {
		t.Errorf("Name = %q, want %q", got, "hnsstat")
	}
	if got := h.URL("hnsstat"); got != "https://hnsscan.com/names/hnsstat" {
		t.Errorf("URL = %q, want %q", got, "https://hnsscan.com/names/hnsstat")
	}
	if !Paid(h) {
		t.Error("Paid = false, want true")
	}
}
