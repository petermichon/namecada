package provider

import (
	"context"
	"io"
	"net/http"
)

// codeMap maps HTTP status codes to a check outcome.
type codeMap struct {
	available []int
	taken     []int
}

// defaultCodes treats 404 as free and 200 as taken, matching the frontend
// checkHttp defaults.
var defaultCodes = codeMap{available: []int{http.StatusNotFound}, taken: []int{http.StatusOK}}

// checkHTTP performs a GET and maps the response status through codes. Any
// status that is neither taken nor available yields StatusUnknown with no
// error; transport failures are returned as errors.
func checkHTTP(ctx context.Context, client *http.Client, url string, codes codeMap) (Result, error) {
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

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
