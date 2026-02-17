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
	//
	// SECURITY: This value must only be used server-side. Never ship it
	// in a mobile app, browser bundle, or any client-facing binary.
	//Load it from an environment variable or secrets manager.
	//
	//Required. Get this from your video library settings in the Bunny dashboard.
	APIKey string

	// LibraryID is the ID of the video library to interact with.
	// Required. Find this in your Bunny dashboard under Stream > Video Library.
	LibraryID string

	// CDNHostname is the pull zone hostname assigned to your video library.
	// Used to construct HLS, thumbnail, preview, and MP4 URLs.
	//
	// Find this in your Bunny dashboard under Stream > Your Library > API.
	// It looks like: "your-zone-name.b-cdn.net"
	//
	// This field is optional, but required for HLS, thumbnail, preview
	// animation, and MP4 URLs. Without it, GetVideoURLs will only return
	// EmbedURL and DirectPlayURL.
	CDNHostname string

	// EmbedTokenKey is the security key for signing iframe embed URLs.
	// Required only when Embed View Token Authentication is enabled in your
	// library's security settings.
	//
	// SECURITY: This value must only be used server-side. Never ship it
	// in a mobile app, browser bundle, or any client-facing binary.
	// Load it from an environment variable or secrets manager.
	//
	// Get this from: Stream Dashboard → Library → Security → Embed View Token Authentication Key.
	//
	// This field is optional.
	EmbedTokenKey string

	// CDNTokenKey is the security key for signing direct CDN URLs (HLS, MP4,
	// thumbnails). Required only when CDN Token Authentication is enabled on
	// your pull zone.
	//
	// SECURITY: This value must only be used server-side. Never ship it
	// in a mobile app, browser bundle, or any client-facing binary.
	// Load it from an environment variable or secrets manager.
	//
	// Get this from: Pull Zone → Security → Token Authentication Key.
	// Note: this is a different key from EmbedTokenKey and lives in a
	// different place in the Bunny dashboard.
	//
	// This field is optional.
	CDNTokenKey string

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
