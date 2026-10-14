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
			var gotURL, gotAccept string
			client := &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					gotURL = req.URL.String()
					gotAccept = req.Header.Get("Accept")
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
			if gotAccept != htmlAccept {
				t.Errorf("Accept = %q, want %q", gotAccept, htmlAccept)
			}
		})
	}
}

// TestNetlifyCheck pins the host-based signal: the request must target
// https://{name}.netlify.app/ and the status must map to taken/available/unknown.
func TestNetlifyCheck(t *testing.T) {
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
			var gotURL, gotAccept string
			client := &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					gotURL = req.URL.String()
					gotAccept = req.Header.Get("Accept")
					return &http.Response{
						StatusCode: tt.code,
						Body:       io.NopCloser(strings.NewReader("")),
						Header:     make(http.Header),
					}, nil
				}),
			}

			e := NewNetlify().(*endpoint)
			e.client = client

			got, err := e.Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if gotURL != "https://acema.netlify.app/" {
				t.Errorf("url = %q, want %q", gotURL, "https://acema.netlify.app/")
			}
			if gotAccept != htmlAccept {
				t.Errorf("Accept = %q, want %q", gotAccept, htmlAccept)
			}
		})
	}
}

// TestFirebaseCheck pins the host-based signal: the request must target
// https://{name}.web.app/ and the status must map to taken/available/unknown.
func TestFirebaseCheck(t *testing.T) {
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
			var gotURL, gotAccept string
			client := &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					gotURL = req.URL.String()
					gotAccept = req.Header.Get("Accept")
					return &http.Response{
						StatusCode: tt.code,
						Body:       io.NopCloser(strings.NewReader("")),
						Header:     make(http.Header),
					}, nil
				}),
			}

			e := NewFirebase().(*endpoint)
			e.client = client

			got, err := e.Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if gotURL != "https://acema.web.app/" {
				t.Errorf("url = %q, want %q", gotURL, "https://acema.web.app/")
			}
			if gotAccept != htmlAccept {
				t.Errorf("Accept = %q, want %q", gotAccept, htmlAccept)
			}
		})
	}
}

// TestNeocitiesCheck pins the host-based signal: the request must target
// https://{name}.neocities.org/ and the status must map to
// taken/available/unknown.
func TestNeocitiesCheck(t *testing.T) {
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
			var gotURL, gotAccept string
			client := &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					gotURL = req.URL.String()
					gotAccept = req.Header.Get("Accept")
					return &http.Response{
						StatusCode: tt.code,
						Body:       io.NopCloser(strings.NewReader("")),
						Header:     make(http.Header),
					}, nil
				}),
			}

			e := NewNeocities().(*endpoint)
			e.client = client

			got, err := e.Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if gotURL != "https://acema.neocities.org/" {
				t.Errorf("url = %q, want %q", gotURL, "https://acema.neocities.org/")
			}
			if gotAccept != htmlAccept {
				t.Errorf("Accept = %q, want %q", gotAccept, htmlAccept)
			}
		})
	}
}

// TestRenderCheck pins the host-based signal: the request must target
// https://{name}.onrender.com/ with the host-status HTML Accept and the status
// must map to taken/available/unknown.
func TestRenderCheck(t *testing.T) {
	tests := []struct {
		name string
		code int
		want Status
	}{
		{"taken", http.StatusOK, StatusTaken},
		{"available", http.StatusNotFound, StatusAvailable},
		{"spun down", http.StatusServiceUnavailable, StatusUnknown},
		{"unknown", http.StatusInternalServerError, StatusUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotURL, gotAccept string
			client := &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					gotURL = req.URL.String()
					gotAccept = req.Header.Get("Accept")
					return &http.Response{
						StatusCode: tt.code,
						Body:       io.NopCloser(strings.NewReader("")),
						Header:     make(http.Header),
					}, nil
				}),
			}

			e := NewRender().(*endpoint)
			e.client = client

			got, err := e.Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if gotURL != "https://acema.onrender.com/" {
				t.Errorf("url = %q, want %q", gotURL, "https://acema.onrender.com/")
			}
			if gotAccept != htmlAccept {
				t.Errorf("Accept = %q, want %q", gotAccept, htmlAccept)
			}
		})
	}
}

// TestRailwayCheck pins the host-based signal: the request must target
// https://{name}.up.railway.app/ with the host-status HTML Accept and the status
// must map to taken/available/unknown.
func TestRailwayCheck(t *testing.T) {
	tests := []struct {
		name string
		code int
		want Status
	}{
		{"taken", http.StatusOK, StatusTaken},
		{"available", http.StatusNotFound, StatusAvailable},
		{"unknown", http.StatusNotAcceptable, StatusUnknown},
		{"unknown server error", http.StatusInternalServerError, StatusUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotURL, gotAccept string
			client := &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					gotURL = req.URL.String()
					gotAccept = req.Header.Get("Accept")
					return &http.Response{
						StatusCode: tt.code,
						Body:       io.NopCloser(strings.NewReader("")),
						Header:     make(http.Header),
					}, nil
				}),
			}

			e := NewRailway().(*endpoint)
			e.client = client

			got, err := e.Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if gotURL != "https://acema.up.railway.app/" {
				t.Errorf("url = %q, want %q", gotURL, "https://acema.up.railway.app/")
			}
			if gotAccept != htmlAccept {
				t.Errorf("Accept = %q, want %q", gotAccept, htmlAccept)
			}
		})
	}
}

// TestHerokuCheck pins the host-based signal: the request must target
// https://{name}.herokuapp.com/ and the status must map to
// taken/available/unknown. Heroku is paid-only, so the paid flag must be set.
func TestHerokuCheck(t *testing.T) {
	e := NewHeroku().(*endpoint)
	if !e.Paid() {
		t.Error("Paid = false, want true")
	}
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
			var gotURL, gotAccept string
			client := &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					gotURL = req.URL.String()
					gotAccept = req.Header.Get("Accept")
					return &http.Response{
						StatusCode: tt.code,
						Body:       io.NopCloser(strings.NewReader("")),
						Header:     make(http.Header),
					}, nil
				}),
			}

			e := NewHeroku().(*endpoint)
			e.client = client

			got, err := e.Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if gotURL != "https://acema.herokuapp.com/" {
				t.Errorf("url = %q, want %q", gotURL, "https://acema.herokuapp.com/")
			}
			if gotAccept != htmlAccept {
				t.Errorf("Accept = %q, want %q", gotAccept, htmlAccept)
			}
		})
	}
}

// TestSurgeCheck pins the host-based signal: the request must target
// https://{name}.surge.sh/ and the status must map to taken/available/unknown.
func TestSurgeCheck(t *testing.T) {
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
			var gotURL, gotAccept string
			client := &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					gotURL = req.URL.String()
					gotAccept = req.Header.Get("Accept")
					return &http.Response{
						StatusCode: tt.code,
						Body:       io.NopCloser(strings.NewReader("")),
						Header:     make(http.Header),
					}, nil
				}),
			}

			e := NewSurge().(*endpoint)
			e.client = client

			got, err := e.Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if gotURL != "https://acema.surge.sh/" {
				t.Errorf("url = %q, want %q", gotURL, "https://acema.surge.sh/")
			}
			if gotAccept != htmlAccept {
				t.Errorf("Accept = %q, want %q", gotAccept, htmlAccept)
			}
		})
	}
}

// TestReadTheDocsCheck pins the host-based signal: the request must target
// https://{name}.readthedocs.io/ and the status must map to
// taken/available/unknown. A claimed project answers 302, which DefaultClient
// follows to 200; an unfollowed 302 is not in the code table, so it classifies
// as unknown.
func TestReadTheDocsCheck(t *testing.T) {
	tests := []struct {
		name string
		code int
		want Status
	}{
		{"taken", http.StatusOK, StatusTaken},
		{"available", http.StatusNotFound, StatusAvailable},
		{"redirect", http.StatusFound, StatusUnknown},
		{"unknown", http.StatusInternalServerError, StatusUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotURL, gotAccept string
			client := &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					gotURL = req.URL.String()
					gotAccept = req.Header.Get("Accept")
					return &http.Response{
						StatusCode: tt.code,
						Body:       io.NopCloser(strings.NewReader("")),
						Header:     make(http.Header),
					}, nil
				}),
			}

			e := NewReadTheDocs().(*endpoint)
			e.client = client

			got, err := e.Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if gotURL != "https://acema.readthedocs.io/" {
				t.Errorf("url = %q, want %q", gotURL, "https://acema.readthedocs.io/")
			}
			if gotAccept != htmlAccept {
				t.Errorf("Accept = %q, want %q", gotAccept, htmlAccept)
			}
		})
	}
}

// TestGumroadCheck pins the host-based signal: the request must target
// https://{name}.gumroad.com/ with the host-status HTML Accept and the status
// must map to taken/available/unknown.
func TestGumroadCheck(t *testing.T) {
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
			var gotURL, gotAccept string
			client := &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					gotURL = req.URL.String()
					gotAccept = req.Header.Get("Accept")
					return &http.Response{
						StatusCode: tt.code,
						Body:       io.NopCloser(strings.NewReader("")),
						Header:     make(http.Header),
					}, nil
				}),
			}

			e := NewGumroad().(*endpoint)
			e.client = client

			got, err := e.Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if gotURL != "https://acema.gumroad.com/" {
				t.Errorf("url = %q, want %q", gotURL, "https://acema.gumroad.com/")
			}
			if gotAccept != htmlAccept {
				t.Errorf("Accept = %q, want %q", gotAccept, htmlAccept)
			}
		})
	}
}

// TestSubstackCheck pins the host-based signal: the request must target
// https://{name}.substack.com/ with the host-status HTML Accept and the status
// must map to taken/available/unknown.
func TestSubstackCheck(t *testing.T) {
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
			var gotURL, gotAccept string
			client := &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					gotURL = req.URL.String()
					gotAccept = req.Header.Get("Accept")
					return &http.Response{
						StatusCode: tt.code,
						Body:       io.NopCloser(strings.NewReader("")),
						Header:     make(http.Header),
					}, nil
				}),
			}

			e := NewSubstack().(*endpoint)
			e.client = client

			got, err := e.Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if gotURL != "https://acema.substack.com/" {
				t.Errorf("url = %q, want %q", gotURL, "https://acema.substack.com/")
			}
			if gotAccept != htmlAccept {
				t.Errorf("Accept = %q, want %q", gotAccept, htmlAccept)
			}
		})
	}
}

// TestShopifyCheck pins the host-based signal: the request must target
// https://{name}.myshopify.com/ with the host-status HTML Accept and the status
// must map to taken/available/unknown. A claimed store answers 301, which
// DefaultClient follows to 200; an unfollowed 301 is not in the code table, so
// it classifies as unknown.
func TestShopifyCheck(t *testing.T) {
	tests := []struct {
		name string
		code int
		want Status
	}{
		{"taken", http.StatusOK, StatusTaken},
		{"available", http.StatusNotFound, StatusAvailable},
		{"redirect", http.StatusMovedPermanently, StatusUnknown},
		{"unknown", http.StatusInternalServerError, StatusUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotURL, gotAccept string
			client := &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					gotURL = req.URL.String()
					gotAccept = req.Header.Get("Accept")
					return &http.Response{
						StatusCode: tt.code,
						Body:       io.NopCloser(strings.NewReader("")),
						Header:     make(http.Header),
					}, nil
				}),
			}

			e := NewShopify().(*endpoint)
			e.client = client

			got, err := e.Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if gotURL != "https://acema.myshopify.com/" {
				t.Errorf("url = %q, want %q", gotURL, "https://acema.myshopify.com/")
			}
			if gotAccept != htmlAccept {
				t.Errorf("Accept = %q, want %q", gotAccept, htmlAccept)
			}
		})
	}
}

// TestDenoDeployCheck pins the host-based signal: the request must target
// https://{name}.deno.dev/ and the status must map to taken/available/unknown.
func TestDenoDeployCheck(t *testing.T) {
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
			var gotURL, gotAccept string
			client := &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					gotURL = req.URL.String()
					gotAccept = req.Header.Get("Accept")
					return &http.Response{
						StatusCode: tt.code,
						Body:       io.NopCloser(strings.NewReader("")),
						Header:     make(http.Header),
					}, nil
				}),
			}

			e := NewDenoDeploy().(*endpoint)
			e.client = client

			got, err := e.Check(context.Background(), "acema")
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if gotURL != "https://acema.deno.dev/" {
				t.Errorf("url = %q, want %q", gotURL, "https://acema.deno.dev/")
			}
			if gotAccept != htmlAccept {
				t.Errorf("Accept = %q, want %q", gotAccept, htmlAccept)
			}
		})
	}
}
