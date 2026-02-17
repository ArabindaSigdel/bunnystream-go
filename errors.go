package bunnystream

import (
	"errors"
	"fmt"
)

// Predefined errors.
var (
	ErrInvalidConfig      = errors.New("invalid config")
	ErrAPIKeyRequired     = errors.New("api key required")
	ErrLibraryIDRequired  = errors.New("library id required")
	ErrInvalidMaxRetries  = errors.New("max retries must be greater than 0")
	ErrInvalidTimeout     = errors.New("timeout must be greater than 0")
	ErrVideoNotFound      = errors.New("video not found")
	ErrUnauthorized       = errors.New("unauthorized - check your API key")
	ErrRateLimited        = errors.New("rate limited - too many requests")
	ErrBadRequest         = errors.New("bad request - check your input")
	ErrInternalServer     = errors.New("internal server error")
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrTitleRequired      = errors.New("title is required")
	ErrVideoIDRequired    = errors.New("video id is required")
)

// APIError represents an error response from the Bunny Stream API.
type APIError struct {
	StatusCode int
	Message    string
	Body       string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("bunny stream api error (status %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("bunny stream api error (status %d)", e.StatusCode)
}

// newAPIError creates a new APIError from status code and response body.
func newAPIError(statusCode int, body []byte) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Body:       string(body),
		Message:    parseErrorMessage(body),
	}
}

// parseErrorMessage attempts to extract error message from response body.
func parseErrorMessage(body []byte) string {
	// For now, just return the raw body
	// TODO: Parse JSON error responses if Bunny returns structured errors
	if len(body) > 200 {
		return string(body[:200]) + "..."
	}
	return string(body)
}
