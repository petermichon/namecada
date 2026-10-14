package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// NewGitHub checks a user or organisation on GitHub. Its minimum length (2) is
// declared in the known table.
func NewGitHub() Provider {
	return newEndpoint("github", "github",
		"https://api.github.com", "/users/%s", "https://github.com/%s", defaultCodes)
}

// NewBitbucket checks a workspace on Bitbucket. The API answers 403 for
// workspaces it knows but hides, so 200 and 403 both mean taken. Its minimum
// length (4) is declared in the known table.
func NewBitbucket() Provider {
	return newEndpoint("bitbucket", "bitbucket",
		"https://api.bitbucket.org", "/2.0/workspaces/%s", "https://bitbucket.org/%s/",
		codeMap{
			available: []int{http.StatusNotFound},
			taken:     []int{http.StatusOK, http.StatusForbidden},
		})
}

// GitLab checks a user or group on gitlab.com. The users API returns a JSON
// array (empty when the name is free), so an empty array falls back to the
// group endpoint, which answers 200 for a taken group and 404 for a free name.
// Source: docs/platforms.md — gitlab.com/api/v4/users?username={name}.
type GitLab struct {
	base
	// BaseURL is the API origin; "https://gitlab.com" when defaulted.
	BaseURL string
	// Client is used for the request; DefaultClient when nil.
	Client *http.Client
}

// NewGitLab returns a provider bound to gitlab.com.
func NewGitLab() *GitLab {
	return &GitLab{
		base:    base{id: "gitlab", platform: "gitlab", urlFmt: "https://gitlab.com/%s"},
		BaseURL: "https://gitlab.com",
		Client:  DefaultClient,
	}
}

// Check queries the users endpoint first; a non-empty array is taken. An empty
// array falls back to the groups endpoint (200 = taken, 404 = available).
func (g *GitLab) Check(ctx context.Context, label string) (Result, error) {
	users := fmt.Sprintf("%s/api/v4/users?username=%s", g.BaseURL, url.QueryEscape(label))
	resp, err := doGet(ctx, g.Client, users)
	if err != nil {
		return Result{}, err
	}
	if resp.StatusCode != http.StatusOK {
		drainAndClose(resp)
		return Result{Status: StatusUnknown}, nil
	}
	var data []json.RawMessage
	decodeErr := json.NewDecoder(resp.Body).Decode(&data)
	drainAndClose(resp)
	if decodeErr != nil {
		return Result{Status: StatusUnknown}, nil
	}
	if len(data) > 0 {
		return Result{Status: StatusTaken}, nil
	}

	group := fmt.Sprintf("%s/api/v4/groups/%s", g.BaseURL, url.PathEscape(label))
	return checkHTTP(ctx, g.Client, group, defaultCodes)
}

// Codeberg checks a user or organisation on Codeberg (Forgejo). It parses the
// 404 body: a free name reports the user "does not exist", while a
// limited-visibility account can 404 with a different body — so that name is
// taken. Source: docs/platforms.md — codeberg.org/api/v1/users/{name}.
type Codeberg struct {
	base
	// BaseURL is the API origin; "https://codeberg.org" when defaulted.
	BaseURL string
	// Client is used for the request; DefaultClient when nil.
	Client *http.Client
}

// NewCodeberg returns a provider bound to codeberg.org.
func NewCodeberg() *Codeberg {
	return &Codeberg{
		base:    base{id: "codeberg", platform: "codeberg", urlFmt: "https://codeberg.org/%s"},
		BaseURL: "https://codeberg.org",
		Client:  DefaultClient,
	}
}

// Check calls the users endpoint: 200 is taken; on a 404 the body message is
// inspected, reporting available only when it mentions the user "does not
// exist". Anything else is unknown.
func (c *Codeberg) Check(ctx context.Context, label string) (Result, error) {
	target := fmt.Sprintf("%s/api/v1/users/%s", c.BaseURL, url.PathEscape(label))
	resp, err := doGet(ctx, c.Client, target)
	if err != nil {
		return Result{}, err
	}
	defer drainAndClose(resp)

	switch resp.StatusCode {
	case http.StatusOK:
		return Result{Status: StatusTaken}, nil
	case http.StatusNotFound:
		var data struct {
			Message *string `json:"message"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&data)
		if data.Message == nil {
			return Result{Status: StatusUnknown}, nil
		}
		if strings.Contains(*data.Message, "does not exist") {
			return Result{Status: StatusAvailable}, nil
		}
		return Result{Status: StatusTaken}, nil
	default:
		return Result{Status: StatusUnknown}, nil
	}
}
