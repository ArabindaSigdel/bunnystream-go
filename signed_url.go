package bunnystream

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

var (
	// ErrEmbedTokenKeyRequired is returned when SignedEmbedURL is called but
	// EmbedTokenKey is not set in Config.
	// Get this from: Stream Dashboard → Library → Security → Embed View Token Authentication Key.
	ErrEmbedTokenKeyRequired = errors.New("embed token key required — set EmbedTokenKey in Config")

	// ErrCDNTokenKeyRequired is returned when a signed CDN URL is requested but
	// CDNTokenKey is not set in Config.
	// Get this from: Pull Zone → Security → Token Authentication Key.
	ErrCDNTokenKeyRequired = errors.New("cdn token key required — set CDNTokenKey in Config")
)

// SignedURLOptions configures optional parameters for signed CDN URLs.
type SignedURLOptions struct {
	// UserIP restricts the signed URL to a specific IPv4 address.
	// By default, the full /24 subnet is allowed to reduce false negatives
	// (e.g. 1.2.3.4 allows 1.2.3.0/24).
	UserIP string

	// CountriesAllowed is a comma-separated list of country codes that may
	// access the URL. All other countries are blocked.
	// Example: "US,GB,DE"
	CountriesAllowed string

	// CountriesBlocked is a comma-separated list of country codes that may
	// NOT access the URL.
	// Example: "CN,RU"
	CountriesBlocked string
}

// SignedURLOption configures optional parameters for signed CDN URLs.
type SignedURLOption func(*SignedURLOptions)

// WithUserIP restricts the signed URL to a specific IPv4 address.
func WithUserIP(ip string) SignedURLOption {
	return func(o *SignedURLOptions) {
		o.UserIP = ip
	}
}

// WithCountriesAllowed restricts access to specific countries.
// Accepts ISO 3166-1 alpha-2 codes, comma-separated.
func WithCountriesAllowed(countries string) SignedURLOption {
	return func(o *SignedURLOptions) {
		o.CountriesAllowed = countries
	}
}

// WithCountriesBlocked blocks access from specific countries.
// Accepts ISO 3166-1 alpha-2 codes, comma-separated.
func WithCountriesBlocked(countries string) SignedURLOption {
	return func(o *SignedURLOptions) {
		o.CountriesBlocked = countries
	}
}

// SignedEmbedURL returns a time-limited signed embed URL for Bunny's iframe player.
//
// Use this when Embed View Token Authentication is enabled in your library's
// security settings, which prevents other websites from hotlinking your player.
//
// The token is a SHA256 hex hash of: EmbedTokenKey + videoID + expiry.
//
// Requires EmbedTokenKey to be set in Config.
// Get this key from: Stream Dashboard → Library → Security → Embed View Token Authentication Key.
//
//	https://iframe.mediadelivery.net/embed/123/video-guid?token=abc123&expires=1234567890
func (c *Client) SignedEmbedURL(videoID string, ttl time.Duration) (string, error) {
	if strings.TrimSpace(videoID) == "" {
		return "", ErrVideoIDRequired
	}
	if c.config.EmbedTokenKey == "" {
		return "", ErrEmbedTokenKeyRequired
	}

	expiry := time.Now().Add(ttl).Unix()
	hash := sha256.Sum256([]byte(c.config.EmbedTokenKey + videoID + fmt.Sprintf("%d", expiry)))
	token := hex.EncodeToString(hash[:])

	base := fmt.Sprintf("https://iframe.mediadelivery.net/embed/%s/%s", c.libraryID, videoID)
	return fmt.Sprintf("%s?token=%s&expires=%d", base, token, expiry), nil
}

// SignedHLSURL returns a time-limited signed HLS playlist URL using a
// directory token.
//
// This is the correct way to sign HLS streams. A regular single-file token
// only covers the playlist itself — all the .ts segment requests it references
// would fail with 403. A directory token signs the entire /{videoID}/ path,
// and the token is embedded in the URL path so browsers automatically
// propagate it to every subsequent chunk request.
//
// Requires both CDNHostname and CDNTokenKey to be set in Config.
// Get the token key from: Pull Zone → Security → Token Authentication Key.
//
//	https://vz-abc.b-cdn.net/bcdn_token=TOKEN&expires=EXP&token_path=/video-guid//video-guid/playlist.m3u8
func (c *Client) SignedHLSURL(videoID string, ttl time.Duration, opts ...SignedURLOption) (string, error) {
	if strings.TrimSpace(videoID) == "" {
		return "", ErrVideoIDRequired
	}
	if c.config.CDNHostname == "" {
		return "", ErrCDNHostnameRequired
	}
	if c.config.CDNTokenKey == "" {
		return "", ErrCDNTokenKeyRequired
	}

	options := &SignedURLOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Sign the directory, not just the file. This covers all .ts chunks too.
	dirPath := fmt.Sprintf("/%s/", videoID)
	filePath := fmt.Sprintf("/%s/playlist.m3u8", videoID)
	expiry := time.Now().Add(ttl).Unix()

	token, err := signCDNToken(c.config.CDNTokenKey, dirPath, expiry, options)
	if err != nil {
		return "", err
	}

	// Path-based token format — browser propagates token to sub-requests automatically.
	host := strings.TrimRight(c.config.CDNHostname, "/")
	signed := fmt.Sprintf("https://%s/bcdn_token=%s&expires=%d&token_path=%s%s",
		host, token, expiry, url.QueryEscape(dirPath), filePath)

	return signed, nil
}

// SignedMP4URL returns a time-limited signed direct MP4 download URL.
//
// Use this when CDN Token Authentication is enabled on your pull zone and
// you need a download link that expires.
//
// Requires both CDNHostname and CDNTokenKey to be set in Config.
// Get the token key from: Pull Zone → Security → Token Authentication Key.
//
//	https://vz-abc.b-cdn.net/video-guid/play_720p.mp4?token=TOKEN&expires=EXP
func (c *Client) SignedMP4URL(videoID string, r Resolution, ttl time.Duration, opts ...SignedURLOption) (string, error) {
	if strings.TrimSpace(videoID) == "" {
		return "", ErrVideoIDRequired
	}
	if r == "" {
		return "", ErrResolutionRequired
	}
	if c.config.CDNHostname == "" {
		return "", ErrCDNHostnameRequired
	}
	if c.config.CDNTokenKey == "" {
		return "", ErrCDNTokenKeyRequired
	}

	options := &SignedURLOptions{}
	for _, opt := range opts {
		opt(options)
	}

	filePath := fmt.Sprintf("/%s/play_%s.mp4", videoID, r)
	expiry := time.Now().Add(ttl).Unix()

	token, err := signCDNToken(c.config.CDNTokenKey, filePath, expiry, options)
	if err != nil {
		return "", err
	}

	host := strings.TrimRight(c.config.CDNHostname, "/")
	base := fmt.Sprintf("https://%s%s", host, filePath)

	params := url.Values{}
	params.Set("token", token)
	params.Set("expires", fmt.Sprintf("%d", expiry))
	if options.CountriesAllowed != "" {
		params.Set("token_countries", options.CountriesAllowed)
	}
	if options.CountriesBlocked != "" {
		params.Set("token_countries_blocked", options.CountriesBlocked)
	}

	return base + "?" + params.Encode(), nil
}

// signCDNToken computes a Bunny CDN Token Authentication V2 token.
//
// Algorithm:
//
//	Base64Encode(SHA256_RAW(key + path + expiry + optionalIP + optionalQueryParams))
//
// Query parameters (excluding "token" and "expires") must be sorted ascending
// and appended as form-encoded key=value pairs (not URL encoded).
func signCDNToken(key, path string, expiry int64, opts *SignedURLOptions) (string, error) {
	// Collect extra query parameters (excluding token and expires).
	extraParams := url.Values{}
	if opts.CountriesAllowed != "" {
		extraParams.Set("token_countries", opts.CountriesAllowed)
	}
	if opts.CountriesBlocked != "" {
		extraParams.Set("token_countries_blocked", opts.CountriesBlocked)
	}

	// Sort keys ascending and build form-encoded string without URL encoding.
	keys := make([]string, 0, len(extraParams))
	for k := range extraParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var paramStr string
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, k+"="+extraParams.Get(k))
	}
	paramStr = strings.Join(pairs, "&")

	// Build the hashable string.
	hashable := key + path + fmt.Sprintf("%d", expiry)
	if opts.UserIP != "" {
		hashable += opts.UserIP
	}
	if paramStr != "" {
		hashable += paramStr
	}

	h := sha256.New()
	h.Write([]byte(hashable))
	raw := h.Sum(nil)

	// Base64 encode and replace characters per Bunny's spec.
	token := base64.StdEncoding.EncodeToString(raw)
	token = strings.ReplaceAll(token, "\n", "")
	token = strings.ReplaceAll(token, "+", "-")
	token = strings.ReplaceAll(token, "/", "_")
	token = strings.ReplaceAll(token, "=", "")

	return token, nil
}
