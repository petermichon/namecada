package provider

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

// roundTripFunc adapts a function to an http.RoundTripper so a test can answer
// a request without a listener.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// TestVercelCheck pins the host-based signal: the request must target
// https://{name}.vercel.app/ and the status must map to taken/available/unknown.
func TestVercelCheck(t *testing.T) {
	tests := []struct {
		name string
		code int
		want Status
	}{
		{"taken", http.StatusOK, StatusTaken},
		{"available", http.StatusNotFound, StatusAvailable},
		{"unknown", http.StatusInternalServerError, StatusUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotURL string
			client := &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					gotURL = req.URL.String()
					return &http.Response{
						StatusCode: tt.code,
						Body:       io.NopCloser(strings.NewReader("")),
						Header:     make(http.Header),
					}, nil
				}),
			}

			e := NewVercel().(*endpoint)
			e.client = client

			got, err := e.Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if gotURL != "https://acema.vercel.app/" {
				t.Errorf("url = %q, want %q", gotURL, "https://acema.vercel.app/")
			}
		})
	}
}
