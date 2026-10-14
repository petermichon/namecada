package provider

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"unicode"
)

// The providers in this file are the credential-gated company-name sources:
// Companies House (UK) and OpenCorporates. Both are legal company-name signals:
// the label is a company name and an exact-name hit on the register means the
// name is taken; anything else is free. Search endpoints are deliberately loose
// on both services, so a hit only counts when the returned name equals the
// queried label — a fuzzy match is not "taken".

// companyQuery maps a scan label to the search term a company register is
// queried with, and normalizes a returned company name so the two can be
// compared. Labels use '-' or '_' as word separators while legal names use
// spaces, so every separator (and any run of whitespace) folds to a single
// space: "acme-corp" queries "acme corp" and matches "ACME  CORP".
func companyQuery(label string) string {
	return strings.Join(strings.FieldsFunc(label, isCompanySeparator), " ")
}

// isCompanySeparator reports whether r separates words in a label or a company
// name.
func isCompanySeparator(r rune) bool {
	return r == '-' || r == '_' || unicode.IsSpace(r)
}

// companyNameMatches reports whether a register hit is an exact match for the
// queried label, comparing case-folded, whitespace-normalized forms.
func companyNameMatches(label, name string) bool {
	return strings.EqualFold(companyQuery(label), companyQuery(name))
}

// basicAuth builds the base64 credential an HTTP Basic Authorization header
// carries for the Companies House key (key as username, empty password).
func basicAuth(user string) string {
	return base64.StdEncoding.EncodeToString([]byte(user + ":"))
}

// CompaniesHouse checks a company name on the UK Companies House register. It
// queries GET api.company-information.service.gov.uk/search/companies?q={name}
// and treats an exact name match as taken; no match (or no items) is free. The
// search is loose, so only an exact hit counts. The API key travels in an HTTP
// Basic Authorization header (key as the username, empty password), never in
// the URL; a 401 (bad key) or any other non-200 is unknown.
type CompaniesHouse struct {
	base
	// Key is the API key (DOMAINBROWSER_COMPANIESHOUSE_KEY).
	Key string
	// BaseURL is the API origin; "https://api.company-information.service.gov.uk"
	// when defaulted.
	BaseURL string
	// Client is used for the request; DefaultClient when nil.
	Client *http.Client
}

// NewCompaniesHouse returns a provider bound to the Companies House API and
// authenticated with key.
func NewCompaniesHouse(key string) *CompaniesHouse {
	return &CompaniesHouse{
		base: base{
			id:       "companieshouse",
			platform: "companieshouse",
			urlFmt:   "https://find-and-update.company-information.service.gov.uk/search/companies?q=%s",
		},
		Key:     key,
		BaseURL: "https://api.company-information.service.gov.uk",
		Client:  DefaultClient,
	}
}

// Check searches the register. A 200 whose items contain the queried label as
// an exact company name is taken; a 200 with no exact hit (or no items) is
// free; any other status, an unreadable body, or invalid JSON is unknown.
func (c *CompaniesHouse) Check(ctx context.Context, label string) (Result, error) {
	query := url.Values{}
	query.Set("q", companyQuery(label))
	target := c.BaseURL + "/search/companies?" + query.Encode()
	resp, err := doGetHeaders(ctx, c.Client, target, map[string]string{
		"Authorization": "Basic " + basicAuth(c.Key),
	})
	if err != nil {
		return Result{}, err
	}
	defer drainAndClose(resp)

	if resp.StatusCode != http.StatusOK {
		return Result{Status: StatusUnknown}, nil
	}

	var data struct {
		Items []struct {
			// The search API reports the company name as "title"; older
			// responses use "company_name". Accept either.
			Title       string `json:"title"`
			CompanyName string `json:"company_name"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return Result{Status: StatusUnknown}, nil
	}
	for _, item := range data.Items {
		name := item.Title
		if name == "" {
			name = item.CompanyName
		}
		if companyNameMatches(label, name) {
			return Result{Status: StatusTaken}, nil
		}
	}
	return Result{Status: StatusAvailable}, nil
}

// OpenCorporates checks a company name across OpenCorporates' open company
// register. It queries GET api.opencorporates.com/v0.4/companies/search?q={name}
// and treats an exact name match as taken; no match is free. The search is
// loose (all words, any order), so only an exact hit counts. The API token
// travels in the X-API-TOKEN header, never in the URL; a 401/403 (bad key or
// quota) or any other non-200 is unknown.
type OpenCorporates struct {
	base
	// Token is the API token (DOMAINBROWSER_OPENCORPORATES_TOKEN).
	Token string
	// BaseURL is the API origin; "https://api.opencorporates.com/v0.4" when
	// defaulted.
	BaseURL string
	// Client is used for the request; DefaultClient when nil.
	Client *http.Client
}

// NewOpenCorporates returns a provider bound to the OpenCorporates API and
// authenticated with token.
func NewOpenCorporates(token string) *OpenCorporates {
	return &OpenCorporates{
		base: base{
			id:       "opencorporates",
			platform: "opencorporates",
			urlFmt:   "https://opencorporates.com/companies?q=%s",
		},
		Token:   token,
		BaseURL: "https://api.opencorporates.com/v0.4",
		Client:  DefaultClient,
	}
}

// Check searches the register. A 200 whose companies contain the queried label
// as an exact company name is taken; a 200 with no exact hit is free; any other
// status, an unreadable body, or invalid JSON is unknown.
func (o *OpenCorporates) Check(ctx context.Context, label string) (Result, error) {
	query := url.Values{}
	query.Set("q", companyQuery(label))
	target := o.BaseURL + "/companies/search?" + query.Encode()
	resp, err := doGetHeaders(ctx, o.Client, target, map[string]string{
		"X-API-TOKEN": o.Token,
	})
	if err != nil {
		return Result{}, err
	}
	defer drainAndClose(resp)

	if resp.StatusCode != http.StatusOK {
		return Result{Status: StatusUnknown}, nil
	}

	var data struct {
		Results struct {
			Companies []struct {
				Company struct {
					Name string `json:"name"`
				} `json:"company"`
			} `json:"companies"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return Result{Status: StatusUnknown}, nil
	}
	for _, c := range data.Results.Companies {
		if companyNameMatches(label, c.Company.Name) {
			return Result{Status: StatusTaken}, nil
		}
	}
	return Result{Status: StatusAvailable}, nil
}
