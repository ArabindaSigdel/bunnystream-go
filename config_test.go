package bunnystream

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

// -----------------------------------------------------------------------------
// validate
// -----------------------------------------------------------------------------

func TestConfig_Validate_MissingAPIKey(t *testing.T) {
	cfg := &Config{LibraryID: "123"}
	err := cfg.validate()

	if !errors.Is(err, ErrAPIKeyRequired) {
		t.Errorf("expected ErrAPIKeyRequired, got %v", err)
	}
	if !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("expected error to wrap ErrInvalidConfig, got %v", err)
	}
}

func TestConfig_Validate_MissingLibraryID(t *testing.T) {
	cfg := &Config{APIKey: "test-key"}
	err := cfg.validate()

	if !errors.Is(err, ErrLibraryIDRequired) {
		t.Errorf("expected ErrLibraryIDRequired, got %v", err)
	}
	if !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("expected error to wrap ErrInvalidConfig, got %v", err)
	}
}

func TestConfig_Validate_BothMissing(t *testing.T) {
	cfg := &Config{}
	err := cfg.validate()

	// validate checks APIKey first, so we expect that error
	if !errors.Is(err, ErrAPIKeyRequired) {
		t.Errorf("expected ErrAPIKeyRequired when both fields missing, got %v", err)
	}
}

func TestConfig_Validate_ValidConfig(t *testing.T) {
	cfg := &Config{APIKey: "test-key", LibraryID: "123"}
	if err := cfg.validate(); err != nil {
		t.Errorf("expected no error for valid config, got %v", err)
	}
}

// -----------------------------------------------------------------------------
// init — defaults
// -----------------------------------------------------------------------------

func TestConfig_Init_SetsDefaultUserAgent(t *testing.T) {
	cfg := &Config{}
	cfg.init()

	if cfg.UserAgent != DefaultUserAgent {
		t.Errorf("UserAgent = %q, want %q", cfg.UserAgent, DefaultUserAgent)
	}
}

func TestConfig_Init_DoesNotOverrideUserAgent(t *testing.T) {
	cfg := &Config{UserAgent: "my-app/1.0"}
	cfg.init()

	if cfg.UserAgent != "my-app/1.0" {
		t.Errorf("UserAgent was overwritten, got %q", cfg.UserAgent)
	}
}

func TestConfig_Init_SetsDefaultMaxRetries(t *testing.T) {
	cfg := &Config{}
	cfg.init()

	if cfg.MaxRetries != DefaultMaxRetries {
		t.Errorf("MaxRetries = %d, want %d", cfg.MaxRetries, DefaultMaxRetries)
	}
}

func TestConfig_Init_DoesNotOverrideMaxRetries(t *testing.T) {
	cfg := &Config{MaxRetries: 5}
	cfg.init()

	if cfg.MaxRetries != 5 {
		t.Errorf("MaxRetries was overwritten, got %d", cfg.MaxRetries)
	}
}

func TestConfig_Init_SetsDefaultTimeout(t *testing.T) {
	cfg := &Config{}
	cfg.init()

	if cfg.Timeout != DefaultTimeout {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, DefaultTimeout)
	}
}

func TestConfig_Init_DoesNotOverrideTimeout(t *testing.T) {
	cfg := &Config{Timeout: 30 * time.Second}
	cfg.init()

	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout was overwritten, got %v", cfg.Timeout)
	}
}

func TestConfig_Init_SetsDefaultBaseURL(t *testing.T) {
	cfg := &Config{}
	cfg.init()

	if cfg.BaseURL != DefaultBaseURL {
		t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, DefaultBaseURL)
	}
}

func TestConfig_Init_DoesNotOverrideBaseURL(t *testing.T) {
	cfg := &Config{BaseURL: "https://custom.api.example.com"}
	cfg.init()

	if cfg.BaseURL != "https://custom.api.example.com" {
		t.Errorf("BaseURL was overwritten, got %q", cfg.BaseURL)
	}
}

func TestConfig_Init_CreatesHTTPClientIfNil(t *testing.T) {
	cfg := &Config{}
	cfg.init()

	if cfg.HTTPClient == nil {
		t.Error("expected HTTPClient to be initialized, got nil")
	}
}

func TestConfig_Init_HTTPClientTimeoutMatchesConfig(t *testing.T) {
	cfg := &Config{Timeout: 10 * time.Second}
	cfg.init()

	if cfg.HTTPClient.Timeout != 10*time.Second {
		t.Errorf("HTTPClient.Timeout = %v, want %v", cfg.HTTPClient.Timeout, 10*time.Second)
	}
}

func TestConfig_Init_DoesNotOverrideHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 5 * time.Second}
	cfg := &Config{HTTPClient: custom}
	cfg.init()

	if cfg.HTTPClient != custom {
		t.Error("HTTPClient was replaced, expected the original custom client")
	}
}

// -----------------------------------------------------------------------------
// NewClient
// -----------------------------------------------------------------------------

func TestNewClient_NilConfig(t *testing.T) {
	_, err := NewClient(nil)
	if !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("expected ErrInvalidConfig for nil config, got %v", err)
	}
}

func TestNewClient_ValidConfig(t *testing.T) {
	cfg := &Config{APIKey: "test-key", LibraryID: "123"}
	client, err := NewClient(cfg)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_ClientFieldsMatchConfig(t *testing.T) {
	cfg := &Config{
		APIKey:    "my-api-key",
		LibraryID: "456",
		BaseURL:   "https://custom.example.com",
	}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.apiKey != "my-api-key" {
		t.Errorf("client.apiKey = %q, want %q", client.apiKey, "my-api-key")
	}
	if client.libraryID != "456" {
		t.Errorf("client.libraryID = %q, want %q", client.libraryID, "456")
	}
	if client.baseURL != "https://custom.example.com" {
		t.Errorf("client.baseURL = %q, want %q", client.baseURL, "https://custom.example.com")
	}
}
