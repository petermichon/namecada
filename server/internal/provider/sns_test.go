package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSNSCheck(t *testing.T) {
	tests := []struct {
		name string
		code int
		body string
		want Status
	}{
		{"registered", http.StatusOK, `{"s":"ok","result":"Fw1ETanDZafof7xEULsnq9UY6o71Tpds89tNwPkWLb1v"}`, StatusTaken},
		{"free", http.StatusOK, `{"s":"error","result":"Domain not found"}`, StatusAvailable},
		{"ok without result", http.StatusOK, `{"s":"ok","result":""}`, StatusUnknown},
		{"unknown envelope", http.StatusOK, `{"s":"weird","result":"?"}`, StatusUnknown},
		{"server error", http.StatusInternalServerError, `{"s":"ok","result":"Fw1ETanDZafof7xEULsnq9UY6o71Tpds89tNwPkWLb1v"}`, StatusUnknown},
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

			s := NewSNS()
			s.BaseURL = srv.URL
			s.Client = srv.Client()

			got, err := s.Check(context.Background(), "bonfida")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if gotPath != "/resolve/bonfida.sol" {
				t.Errorf("path = %q, want %q", gotPath, "/resolve/bonfida.sol")
			}
		})
	}
}

func TestSNSCheckTransportError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close() // no listener: connection refused

	s := NewSNS()
	s.BaseURL = url
	s.Client = srv.Client()
	if _, err := s.Check(context.Background(), "bonfida"); err == nil {
		t.Fatal("expected a transport error, got nil")
	}
}

func TestSNSMetadata(t *testing.T) {
	s := NewSNS()
	if s.ID() != "sns" {
		t.Errorf("ID = %q, want %q", s.ID(), "sns")
	}
	if s.Platform() != "sns" {
		t.Errorf("Platform = %q, want %q", s.Platform(), "sns")
	}
	if got := s.Name("bonfida"); got != "bonfida.sol" {
		t.Errorf("Name = %q, want %q", got, "bonfida.sol")
	}
	if got := s.URL("bonfida"); got != "https://www.sns.id/domain/bonfida.sol" {
		t.Errorf("URL = %q, want %q", got, "https://www.sns.id/domain/bonfida.sol")
	}
	if !Paid(s) {
		t.Error("Paid = false, want true")
	}
}
