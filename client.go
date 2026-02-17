package bunnystream

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client is the Bunny Stream API client.
type Client struct {
	config     *Config
	httpClient *http.Client
	baseURL    string
	libraryID  string
	apiKey     string
}

// NewClient creates a new Bunny Stream client.
// Returns an error if the configuration is invalid.
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, ErrInvalidConfig
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	cfg.init()

	return &Client{
		config:     cfg,
		httpClient: cfg.HTTPClient,
		baseURL:    cfg.BaseURL,
		libraryID:  cfg.LibraryID,
		apiKey:     cfg.APIKey,
	}, nil
}

// request return a http request with all the parms set
func (c *Client) request(ctx context.Context, method, url string, body io.Reader, contentType string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("AccessKey", c.apiKey)
	req.Header.Set("User-Agent", c.config.UserAgent)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return req, nil
}

// doRequest performs an HTTP request and returns the response.
func (c *Client) doRequest(req *http.Request) (*Response, error) {
	// Perform request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	// Create response wrapper
	response, err := newResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for errors
	if err := c.checkResponseError(response.StatusCode, response.Body); err != nil {
		return response, err
	}

	return response, nil
}

// checkResponseError checks if the response indicates an error.
func (c *Client) checkResponseError(statusCode int, body []byte) error {
	switch statusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent:
		return nil
	case http.StatusBadRequest:
		return newAPIError(statusCode, body)
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusNotFound:
		return ErrVideoNotFound
	case http.StatusTooManyRequests:
		return ErrRateLimited
	case http.StatusInternalServerError:
		return ErrInternalServer
	case http.StatusServiceUnavailable:
		return ErrServiceUnavailable
	case http.StatusForbidden:
		return ErrForbidden
	default:
		return newAPIError(statusCode, body)
	}
}

// buildURL constructs a full URL from the endpoint format and arguments.
func (c *Client) buildURL(format string, args ...interface{}) string {
	path := fmt.Sprintf(format, args...)
	return c.baseURL + path
}

// encodeJSON encodes a value to JSON.
func (c *Client) encodeJSON(v interface{}) (io.Reader, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		return nil, fmt.Errorf("failed to encode JSON: %w", err)
	}
	return &buf, nil
}

// decodeJSON decodes JSON from a byte slice.
func (c *Client) decodeJSON(data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}
	return nil
}
