package bunnystream

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// Default values for the Config struct.
const (
	DefaultMaxRetries int           = 3
	DefaultTimeout    time.Duration = 60 * time.Second
	DefaultUserAgent  string        = "bunnystream-go/0.1.0"
	DefaultBaseURL    string        = "https://video.bunnycdn.com"
)

// Config holds the configuration for the Bunny Stream client.
type Config struct {
	// Logger is the structured logger to use for logging information about API
	// requests and responses.
	//
	// This field is optional.
	Logger *slog.Logger

	// APIKey is the API key for authenticating with Bunny Stream.
	// Required. Get this from your video library settings in the Bunny dashboard.
	APIKey string

	// LibraryID is the ID of the video library to interact with.
	// Required. Find this in your Bunny dashboard under Stream > Video Library.
	LibraryID string

	// UserAgent is the user agent to use when making HTTP requests to the API.
	//
	// This field is optional.
	UserAgent string

	// BaseURL is the base URL for the Bunny Stream API.
	//
	// This field is optional. Defaults to DefaultBaseURL if not set.
	BaseURL string

	// HTTPClient is the HTTP client to use for requests.
	//
	// This field is optional. If nil, a default client with DefaultTimeout will be used.
	HTTPClient *http.Client

	// MaxRetries specifies the maximum number of times to retry a request if it
	// fails due to rate limiting or temporary errors.
	//
	// This field is optional. Defaults to DefaultMaxRetries.
	MaxRetries int

	// Timeout is the time limit for requests made by the client to the API.
	//
	// This field is optional. Defaults to DefaultTimeout.
	Timeout time.Duration

	// mu protects Config initialization.
	mu sync.Mutex
}

// init initializes missing Config fields with their default values.
func (c *Config) init() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.UserAgent == "" {
		c.UserAgent = DefaultUserAgent
	}

	if c.MaxRetries < 1 {
		c.MaxRetries = DefaultMaxRetries
	}

	if c.Timeout < 1 {
		c.Timeout = DefaultTimeout
	}

	if c.BaseURL == "" {
		c.BaseURL = DefaultBaseURL
	}

	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{
			Timeout: c.Timeout,
		}
	}
}

// validate returns an error if the config is invalid.
func (c *Config) validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("%w: %w", ErrInvalidConfig, ErrAPIKeyRequired)
	}

	if c.LibraryID == "" {
		return fmt.Errorf("%w: %w", ErrInvalidConfig, ErrLibraryIDRequired)
	}

	return nil
}
