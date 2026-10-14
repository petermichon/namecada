package provider

import (
	"context"
	"io"
	"net/http"
	"time"
)

// DefaultTimeout is the request timeout on DefaultClient. It bounds a single
// source's HTTP call so a hung server cannot stall a scan indefinitely, which
// the timeout-less http.DefaultClient would allow.
const DefaultTimeout = 10 * time.Second

// DefaultClient is the request client providers use when none is supplied. Its
// Timeout applies to the whole request, including reading the body.
var DefaultClient = &http.Client{Timeout: DefaultTimeout}

// codeMap maps HTTP status codes to a check outcome.
type codeMap struct {
	available []int
	taken     []int
}

// defaultCodes treats 404 as free and 200 as taken, matching the frontend
// checkHttp defaults.
var defaultCodes = codeMap{available: []int{http.StatusNotFound}, taken: []int{http.StatusOK}}

// doGet performs a GET with the shared Accept header; a nil client falls back
// to DefaultClient so every request carries a timeout.
func doGet(ctx context.Context, client *http.Client, url string) (*http.Response, error) {
	if client == nil {
		client = DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	return client.Do(req)
}

// drainAndClose fully reads and closes a response body so the connection can be
// reused.
func drainAndClose(resp *http.Response) {
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}

// checkHTTP performs a GET and maps the response status through codes. Any
// status that is neither taken nor available yields StatusUnknown with no
// error; transport failures are returned as errors.
func checkHTTP(ctx context.Context, client *http.Client, url string, codes codeMap) (Result, error) {
	resp, err := doGet(ctx, client, url)
	if err != nil {
		return Result{}, err
	}
	defer drainAndClose(resp)

	return codes.classify(resp.StatusCode), nil
}

func (c codeMap) classify(status int) Result {
	for _, code := range c.taken {
		if status == code {
			return Result{Status: StatusTaken}
		}
	}
	for _, code := range c.available {
		if status == code {
			return Result{Status: StatusAvailable}
		}
	}
	return Result{Status: StatusUnknown}
}
