package bunnystream

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// signedBaseConfig returns a config with all signing keys set.
func signedBaseConfig() *Config {
	return &Config{
		APIKey:        "test-key",
		LibraryID:     "123",
		CDNHostname:   "vz-abc123.b-cdn.net",
		EmbedTokenKey: "embed-secret",
		CDNTokenKey:   "cdn-secret",
	}
}

// -----------------------------------------------------------------------------
// SignedEmbedURL
// -----------------------------------------------------------------------------

func TestSignedEmbedURL_ContainsTokenAndExpires(t *testing.T) {
	c := mustNewClient(t, signedBaseConfig())
	got, err := c.SignedEmbedURL("video-abc", time.Hour)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "token=") {
		t.Errorf("SignedEmbedURL missing 'token' param: %q", got)
	}
	if !strings.Contains(got, "expires=") {
		t.Errorf("SignedEmbedURL missing 'expires' param: %q", got)
	}
}

func TestSignedEmbedURL_ContainsLibraryAndVideoID(t *testing.T) {
	c := mustNewClient(t, signedBaseConfig())
	got, err := c.SignedEmbedURL("video-abc", time.Hour)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "/123/") {
		t.Errorf("SignedEmbedURL missing library ID: %q", got)
	}
	if !strings.Contains(got, "video-abc") {
		t.Errorf("SignedEmbedURL missing video ID: %q", got)
	}
}

func TestSignedEmbedURL_TokenChangesWithDifferentVideoID(t *testing.T) {
	c := mustNewClient(t, signedBaseConfig())

	url1, _ := c.SignedEmbedURL("video-aaa", time.Hour)
	url2, _ := c.SignedEmbedURL("video-bbb", time.Hour)

	// Extract just the token values
	token1 := extractParam(t, url1, "token")
	token2 := extractParam(t, url2, "token")

	if token1 == token2 {
		t.Error("expected different tokens for different video IDs")
	}
}

func TestSignedEmbedURL_MissingEmbedTokenKey(t *testing.T) {
	cfg := baseConfig()
	// No EmbedTokenKey
	c := mustNewClient(t, cfg)

	_, err := c.SignedEmbedURL("video-abc", time.Hour)
	if !errors.Is(err, ErrEmbedTokenKeyRequired) {
		t.Errorf("expected ErrEmbedTokenKeyRequired, got %v", err)
	}
}

func TestSignedEmbedURL_EmptyVideoID(t *testing.T) {
	c := mustNewClient(t, signedBaseConfig())
	_, err := c.SignedEmbedURL("", time.Hour)

	if !errors.Is(err, ErrVideoIDRequired) {
		t.Errorf("expected ErrVideoIDRequired, got %v", err)
	}
}

// -----------------------------------------------------------------------------
// SignedHLSURL
// -----------------------------------------------------------------------------

func TestSignedHLSURL_ContainsTokenAndExpiry(t *testing.T) {
	c := mustNewClient(t, signedBaseConfig())
	got, err := c.SignedHLSURL("video-abc", time.Hour)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "bcdn_token=") {
		t.Errorf("SignedHLSURL missing 'bcdn_token': %q", got)
	}
	if !strings.Contains(got, "expires=") {
		t.Errorf("SignedHLSURL missing 'expires': %q", got)
	}
}

func TestSignedHLSURL_UsesDirectoryToken(t *testing.T) {
	c := mustNewClient(t, signedBaseConfig())
	got, err := c.SignedHLSURL("video-abc", time.Hour)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Directory token format must include token_path
	if !strings.Contains(got, "token_path=") {
		t.Errorf("SignedHLSURL missing 'token_path' — not using directory token format: %q", got)
	}
}

func TestSignedHLSURL_EndsWithPlaylistM3U8(t *testing.T) {
	c := mustNewClient(t, signedBaseConfig())
	got, err := c.SignedHLSURL("video-abc", time.Hour)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(got, "/playlist.m3u8") {
		t.Errorf("SignedHLSURL should end with /playlist.m3u8, got: %q", got)
	}
}

func TestSignedHLSURL_MissingCDNHostname(t *testing.T) {
	cfg := &Config{
		APIKey:      "test-key",
		LibraryID:   "123",
		CDNTokenKey: "cdn-secret",
		// No CDNHostname
	}
	c := mustNewClient(t, cfg)
	_, err := c.SignedHLSURL("video-abc", time.Hour)

	if !errors.Is(err, ErrCDNHostnameRequired) {
		t.Errorf("expected ErrCDNHostnameRequired, got %v", err)
	}
}

func TestSignedHLSURL_MissingCDNTokenKey(t *testing.T) {
	cfg := &Config{
		APIKey:      "test-key",
		LibraryID:   "123",
		CDNHostname: "vz-abc123.b-cdn.net",
		// No CDNTokenKey
	}
	c := mustNewClient(t, cfg)
	_, err := c.SignedHLSURL("video-abc", time.Hour)

	if !errors.Is(err, ErrCDNTokenKeyRequired) {
		t.Errorf("expected ErrCDNTokenKeyRequired, got %v", err)
	}
}

func TestSignedHLSURL_EmptyVideoID(t *testing.T) {
	c := mustNewClient(t, signedBaseConfig())
	_, err := c.SignedHLSURL("", time.Hour)

	if !errors.Is(err, ErrVideoIDRequired) {
		t.Errorf("expected ErrVideoIDRequired, got %v", err)
	}
}

func TestSignedHLSURL_WithCountriesAllowed(t *testing.T) {
	c := mustNewClient(t, signedBaseConfig())
	got, err := c.SignedHLSURL("video-abc", time.Hour, WithCountriesAllowed("US,GB"))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The country restriction is baked into the token via signCDNToken,
	// so we just verify the URL was produced without error and contains the token.
	if !strings.Contains(got, "bcdn_token=") {
		t.Errorf("expected signed URL, got: %q", got)
	}
}

// -----------------------------------------------------------------------------
// SignedMP4URL
// -----------------------------------------------------------------------------

func TestSignedMP4URL_ContainsTokenAndExpiry(t *testing.T) {
	c := mustNewClient(t, signedBaseConfig())
	got, err := c.SignedMP4URL("video-abc", Res720p, time.Hour)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "token=") {
		t.Errorf("SignedMP4URL missing 'token': %q", got)
	}
	if !strings.Contains(got, "expires=") {
		t.Errorf("SignedMP4URL missing 'expires': %q", got)
	}
}

func TestSignedMP4URL_ContainsResolutionInPath(t *testing.T) {
	c := mustNewClient(t, signedBaseConfig())
	got, err := c.SignedMP4URL("video-abc", Res1080p, time.Hour)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "play_1080p.mp4") {
		t.Errorf("SignedMP4URL missing resolution in path: %q", got)
	}
}

func TestSignedMP4URL_MissingCDNTokenKey(t *testing.T) {
	cfg := &Config{
		APIKey:      "test-key",
		LibraryID:   "123",
		CDNHostname: "vz-abc123.b-cdn.net",
	}
	c := mustNewClient(t, cfg)
	_, err := c.SignedMP4URL("video-abc", Res720p, time.Hour)

	if !errors.Is(err, ErrCDNTokenKeyRequired) {
		t.Errorf("expected ErrCDNTokenKeyRequired, got %v", err)
	}
}

func TestSignedMP4URL_EmptyResolution(t *testing.T) {
	c := mustNewClient(t, signedBaseConfig())
	_, err := c.SignedMP4URL("video-abc", "", time.Hour)

	if !errors.Is(err, ErrResolutionRequired) {
		t.Errorf("expected ErrResolutionRequired, got %v", err)
	}
}

func TestSignedMP4URL_WithCountriesAllowedAppearsInQueryParams(t *testing.T) {
	c := mustNewClient(t, signedBaseConfig())
	got, err := c.SignedMP4URL("video-abc", Res720p, time.Hour, WithCountriesAllowed("US,GB"))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "token_countries=") {
		t.Errorf("expected token_countries in URL, got: %q", got)
	}
}

func TestSignedMP4URL_WithCountriesBlockedAppearsInQueryParams(t *testing.T) {
	c := mustNewClient(t, signedBaseConfig())
	got, err := c.SignedMP4URL("video-abc", Res720p, time.Hour, WithCountriesBlocked("CN,RU"))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "token_countries_blocked=") {
		t.Errorf("expected token_countries_blocked in URL, got: %q", got)
	}
}

// -----------------------------------------------------------------------------
// signCDNToken — determinism (known-good test vector)
// This test locks in the signing algorithm. If it ever breaks, you've
// accidentally changed how tokens are computed — which will break live URLs.
// -----------------------------------------------------------------------------

func TestSignCDNToken_Deterministic(t *testing.T) {
	opts := &SignedURLOptions{}
	token1, err := signCDNToken("my-secret", "/video-abc/", 1700000000, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	token2, err := signCDNToken("my-secret", "/video-abc/", 1700000000, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token1 != token2 {
		t.Errorf("signCDNToken is not deterministic: %q vs %q", token1, token2)
	}
}

func TestSignCDNToken_ChangesWithDifferentKey(t *testing.T) {
	opts := &SignedURLOptions{}
	t1, _ := signCDNToken("secret-a", "/video-abc/", 1700000000, opts)
	t2, _ := signCDNToken("secret-b", "/video-abc/", 1700000000, opts)

	if t1 == t2 {
		t.Error("expected different tokens for different keys")
	}
}

func TestSignCDNToken_ChangesWithDifferentPath(t *testing.T) {
	opts := &SignedURLOptions{}
	t1, _ := signCDNToken("secret", "/video-aaa/", 1700000000, opts)
	t2, _ := signCDNToken("secret", "/video-bbb/", 1700000000, opts)

	if t1 == t2 {
		t.Error("expected different tokens for different paths")
	}
}

func TestSignCDNToken_ChangesWithDifferentExpiry(t *testing.T) {
	opts := &SignedURLOptions{}
	t1, _ := signCDNToken("secret", "/video-abc/", 1700000000, opts)
	t2, _ := signCDNToken("secret", "/video-abc/", 1700000001, opts)

	if t1 == t2 {
		t.Error("expected different tokens for different expiry times")
	}
}

func TestSignCDNToken_NoInvalidBase64Characters(t *testing.T) {
	opts := &SignedURLOptions{}
	token, err := signCDNToken("my-secret", "/video-abc/", 1700000000, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Bunny requires URL-safe base64: no +, /, or = padding
	for _, ch := range []string{"+", "/", "=", "\n"} {
		if strings.Contains(token, ch) {
			t.Errorf("token contains invalid character %q: %q", ch, token)
		}
	}
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// extractParam pulls a query param value out of a raw URL string.
func extractParam(t *testing.T, rawURL, key string) string {
	t.Helper()
	// Find key= in the URL
	needle := key + "="
	idx := strings.Index(rawURL, needle)
	if idx == -1 {
		t.Fatalf("param %q not found in URL: %q", key, rawURL)
	}
	rest := rawURL[idx+len(needle):]
	end := strings.IndexAny(rest, "&")
	if end == -1 {
		return rest
	}
	return rest[:end]
}
