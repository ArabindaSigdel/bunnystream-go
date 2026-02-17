package bunnystream

import (
	"io"
	"net/http"
)

// Response represents a response from the Bunny Stream API.
// It includes the HTTP status code, headers, and raw response body
// for debugging and inspection purposes.
type Response struct {
	// StatusCode is the HTTP status code of the response.
	StatusCode int

	// Headers contains the HTTP response headers.
	Headers http.Header

	// Body contains the raw response body.
	Body []byte
}

// newResponse creates a new Response from an HTTP response.
func newResponse(resp *http.Response) (*Response, error) {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       respBody,
	}, nil
}
