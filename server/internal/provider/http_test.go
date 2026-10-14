package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestCheckHTTPChallenge proves a bot-challenge interstitial served with a 200
// is classified as unknown (never cached) rather than taken for every label,
// whether the signal is the cf-mitigated header or a body marker. A real 200
// taken answer is left untouched.
func TestCheckHTTPChallenge(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		headers map[string]string
		body    string
		want    Status
	}{
		{
			name:   "normal 200 taken",
			status: http.StatusOK,
			body:   `{"ok":true}`,
			want:   StatusTaken,
		},
		{
			name:    "cf-mitigated header",
			status:  http.StatusOK,
			headers: map[string]string{"cf-mitigated": "challenge"},
			body:    `<html><body>anything</body></html>`,
			want:    StatusUnknown,
		},
		{
			name:   "client challenge body",
			status: http.StatusOK,
			body:   `<html><head><title>Client Challenge</title></head></html>`,
			want:   StatusUnknown,
		},
		{
			name:   "just a moment body",
			status: http.StatusOK,
			body:   `<html><head><title>Just a moment...</title></head></html>`,
			want:   StatusUnknown,
		},
		{
			name:   "marker on a 404 stays available",
			status: http.StatusNotFound,
			body:   `<html><head><title>Just a moment...</title></head></html>`,
			want:   StatusAvailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				for k, v := range tt.headers {
					w.Header().Set(k, v)
				}
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			got, err := checkHTTP(context.Background(), srv.Client(), srv.URL, defaultCodes)
			if err != nil {
				t.Fatalf("checkHTTP returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
		})
	}
}
