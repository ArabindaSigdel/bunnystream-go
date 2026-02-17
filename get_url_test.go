package bunnystream

import (
	"errors"
	"strings"
	"testing"
)

// mustNewClient creates a client for URL tests. Uses a fake httptest server
// URL as BaseURL so no real network calls are made.
func mustNewClient(t *testing.T, cfg *Config) *Client {
	t.Helper()
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("mustNewClient: %v", err)
	}
	return client
}

// baseConfig returns a minimal valid config for URL tests.
func baseConfig() *Config {
	return &Config{
		APIKey:    "test-key",
		LibraryID: "123",
	}
}

// -----------------------------------------------------------------------------
// EmbedURL
// -----------------------------------------------------------------------------

func TestEmbedURL_ReturnsCorrectURL(t *testing.T) {
	c := mustNewClient(t, baseConfig())
	got, err := c.EmbedURL("video-abc")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "https://iframe.mediadelivery.net/embed/123/video-abc"
	if got != want {
		t.Errorf("EmbedURL = %q, want %q", got, want)
	}
}

func TestEmbedURL_EmptyVideoID(t *testing.T) {
	c := mustNewClient(t, baseConfig())
	_, err := c.EmbedURL("")

	if !errors.Is(err, ErrVideoIDRequired) {
		t.Errorf("expected ErrVideoIDRequired, got %v", err)
	}
}

func TestEmbedURL_WhitespaceVideoID(t *testing.T) {
	c := mustNewClient(t, baseConfig())
	_, err := c.EmbedURL("   ")

	if !errors.Is(err, ErrVideoIDRequired) {
		t.Errorf("expected ErrVideoIDRequired for whitespace videoID, got %v", err)
	}
}

func TestEmbedURL_DoesNotRequireCDNHostname(t *testing.T) {
	cfg := baseConfig()
	// Intentionally no CDNHostname
	c := mustNewClient(t, cfg)
	_, err := c.EmbedURL("video-abc")

	if err != nil {
		t.Errorf("EmbedURL should not require CDNHostname, got error: %v", err)
	}
}

// -----------------------------------------------------------------------------
// DirectPlayURL
// -----------------------------------------------------------------------------

func TestDirectPlayURL_ReturnsCorrectURL(t *testing.T) {
	c := mustNewClient(t, baseConfig())
	got, err := c.DirectPlayURL("video-abc")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "https://video.bunnycdn.com/play/123/video-abc"
	if got != want {
		t.Errorf("DirectPlayURL = %q, want %q", got, want)
	}
}

func TestDirectPlayURL_EmptyVideoID(t *testing.T) {
	c := mustNewClient(t, baseConfig())
	_, err := c.DirectPlayURL("")

	if !errors.Is(err, ErrVideoIDRequired) {
		t.Errorf("expected ErrVideoIDRequired, got %v", err)
	}
}

// -----------------------------------------------------------------------------
// HLSPlaylistURL
// -----------------------------------------------------------------------------

func TestHLSPlaylistURL_ReturnsCorrectURL(t *testing.T) {
	cfg := baseConfig()
	cfg.CDNHostname = "vz-abc123.b-cdn.net"
	c := mustNewClient(t, cfg)

	got, err := c.HLSPlaylistURL("video-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "https://vz-abc123.b-cdn.net/video-abc/playlist.m3u8"
	if got != want {
		t.Errorf("HLSPlaylistURL = %q, want %q", got, want)
	}
}

func TestHLSPlaylistURL_MissingCDNHostname(t *testing.T) {
	c := mustNewClient(t, baseConfig())
	_, err := c.HLSPlaylistURL("video-abc")

	if !errors.Is(err, ErrCDNHostnameRequired) {
		t.Errorf("expected ErrCDNHostnameRequired, got %v", err)
	}
}

func TestHLSPlaylistURL_EmptyVideoID(t *testing.T) {
	cfg := baseConfig()
	cfg.CDNHostname = "vz-abc123.b-cdn.net"
	c := mustNewClient(t, cfg)

	_, err := c.HLSPlaylistURL("")
	if !errors.Is(err, ErrVideoIDRequired) {
		t.Errorf("expected ErrVideoIDRequired, got %v", err)
	}
}

// -----------------------------------------------------------------------------
// ThumbnailURL
// -----------------------------------------------------------------------------

func TestThumbnailURL_ReturnsCorrectURL(t *testing.T) {
	cfg := baseConfig()
	cfg.CDNHostname = "vz-abc123.b-cdn.net"
	c := mustNewClient(t, cfg)

	got, err := c.ThumbnailURL("video-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "https://vz-abc123.b-cdn.net/video-abc/thumbnail.jpg"
	if got != want {
		t.Errorf("ThumbnailURL = %q, want %q", got, want)
	}
}

func TestThumbnailURL_MissingCDNHostname(t *testing.T) {
	c := mustNewClient(t, baseConfig())
	_, err := c.ThumbnailURL("video-abc")

	if !errors.Is(err, ErrCDNHostnameRequired) {
		t.Errorf("expected ErrCDNHostnameRequired, got %v", err)
	}
}

// -----------------------------------------------------------------------------
// PreviewAnimationURL
// -----------------------------------------------------------------------------

func TestPreviewAnimationURL_ReturnsCorrectURL(t *testing.T) {
	cfg := baseConfig()
	cfg.CDNHostname = "vz-abc123.b-cdn.net"
	c := mustNewClient(t, cfg)

	got, err := c.PreviewAnimationURL("video-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "https://vz-abc123.b-cdn.net/video-abc/preview.webp"
	if got != want {
		t.Errorf("PreviewAnimationURL = %q, want %q", got, want)
	}
}

func TestPreviewAnimationURL_MissingCDNHostname(t *testing.T) {
	c := mustNewClient(t, baseConfig())
	_, err := c.PreviewAnimationURL("video-abc")

	if !errors.Is(err, ErrCDNHostnameRequired) {
		t.Errorf("expected ErrCDNHostnameRequired, got %v", err)
	}
}

// -----------------------------------------------------------------------------
// MP4URL
// -----------------------------------------------------------------------------

func TestMP4URL_ReturnsCorrectURL(t *testing.T) {
	cfg := baseConfig()
	cfg.CDNHostname = "vz-abc123.b-cdn.net"
	c := mustNewClient(t, cfg)

	got, err := c.MP4URL("video-abc", Res720p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "https://vz-abc123.b-cdn.net/video-abc/play_720p.mp4"
	if got != want {
		t.Errorf("MP4URL = %q, want %q", got, want)
	}
}

func TestMP4URL_AllResolutions(t *testing.T) {
	cfg := baseConfig()
	cfg.CDNHostname = "vz-abc123.b-cdn.net"
	c := mustNewClient(t, cfg)

	resolutions := []Resolution{Res240p, Res360p, Res480p, Res720p, Res1080p, Res1440p, Res2160p}
	for _, r := range resolutions {
		t.Run(string(r), func(t *testing.T) {
			got, err := c.MP4URL("video-abc", r)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(got, string(r)) {
				t.Errorf("MP4URL %q does not contain resolution %q", got, r)
			}
		})
	}
}

func TestMP4URL_EmptyResolution(t *testing.T) {
	cfg := baseConfig()
	cfg.CDNHostname = "vz-abc123.b-cdn.net"
	c := mustNewClient(t, cfg)

	_, err := c.MP4URL("video-abc", "")
	if !errors.Is(err, ErrResolutionRequired) {
		t.Errorf("expected ErrResolutionRequired, got %v", err)
	}
}

func TestMP4URL_MissingCDNHostname(t *testing.T) {
	c := mustNewClient(t, baseConfig())
	_, err := c.MP4URL("video-abc", Res720p)

	if !errors.Is(err, ErrCDNHostnameRequired) {
		t.Errorf("expected ErrCDNHostnameRequired, got %v", err)
	}
}

func TestMP4URL_EmptyVideoID(t *testing.T) {
	cfg := baseConfig()
	cfg.CDNHostname = "vz-abc123.b-cdn.net"
	c := mustNewClient(t, cfg)

	_, err := c.MP4URL("", Res720p)
	if !errors.Is(err, ErrVideoIDRequired) {
		t.Errorf("expected ErrVideoIDRequired, got %v", err)
	}
}

// -----------------------------------------------------------------------------
// cdnBase — trailing slash handling
// -----------------------------------------------------------------------------

func TestCDNBase_TrailingSlashOnHostname(t *testing.T) {
	cfg := baseConfig()
	cfg.CDNHostname = "vz-abc123.b-cdn.net/" // trailing slash
	c := mustNewClient(t, cfg)

	got, err := c.ThumbnailURL("video-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should not produce double slash
	if strings.Contains(got, "//video-abc") {
		t.Errorf("URL contains double slash: %q", got)
	}
}
