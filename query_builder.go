package bunnystream

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// queryBuilder constructs URL query parameters and applies them to a request.
//
// Usage:
//
//	buildQuery(req).
//	    SetBool("jitEnabled", opts.jitEnabled).
//	    SetStrings("enabledResolutions", opts.enabledResolution).
//	    SetString("sourceLanguage", opts.sourceLanguage).
//	    Apply()
type queryBuilder struct {
	req    *http.Request
	values url.Values
}

// buildQuery creates a new queryBuilder for the given request.
func buildQuery(req *http.Request) *queryBuilder {
	return &queryBuilder{
		req:    req,
		values: req.URL.Query(),
	}
}

// setBool adds a boolean query param if the pointer is non-nil.
func (q *queryBuilder) setBool(key string, v *bool) *queryBuilder {
	if v != nil {
		q.values.Set(key, strconv.FormatBool(*v))
	}
	return q
}

// setString adds a string query param if the value is non-empty.
func (q *queryBuilder) setString(key, v string) *queryBuilder {
	if strings.TrimSpace(v) != "" {
		q.values.Set(key, v)
	}
	return q
}

// setStrings joins a string slice with commas and adds it as a query param
// if the slice is non-empty.
func (q *queryBuilder) setStrings(key string, v []string) *queryBuilder {
	if len(v) > 0 {
		q.values.Set(key, strings.Join(v, ","))
	}
	return q
}

// apply writes the built query params back to the request URL.
// Always call this at the end of the chain — without it, nothing is applied.
func (q *queryBuilder) apply() {
	q.req.URL.RawQuery = q.values.Encode()
}
