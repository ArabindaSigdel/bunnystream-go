package bunnystream

import (
	"errors"
	"fmt"
	"strings"
)

// ErrCDNHostnameRequired is returned when a method requires CDNHostname to be
// set in Config but it isn't. Set it to the pull zone hostname found in your
// Bunny dashboard under Stream > Your Library > API (e.g. "vz-abc123.b-cdn.net").
var ErrCDNHostnameRequired = errors.New("cdn hostname required — set CDNHostname in Config")

// cdnBase returns the base CDN URL for a video, or an error if CDNHostname
// is not configured. Used internally by all CDN-backed URL methods.
func (c *Client) cdnBase(videoID string) (string, error) {
	if c.config.CDNHostname == "" {
		return "", ErrCDNHostnameRequired
	}
	host := strings.TrimRight(c.config.CDNHostname, "/")
	return fmt.Sprintf("https://%s/%s", host, videoID), nil
}

// EmbedURL returns the URL for Bunny's hosted iframe player.
//
// Use this when you want to drop a <iframe src="..."> on a webpage and let
// Bunny handle everything — adaptive quality, controls, captions, and
// cross-browser compatibility.
//
// Only requires LibraryID and the video ID. CDNHostname is not needed.
//
//	<iframe src="https://iframe.mediadelivery.net/embed/123/video-guid" />
func (c *Client) EmbedURL(videoID string) (string, error) {
	if strings.TrimSpace(videoID) == "" {
		return "", ErrVideoIDRequired
	}
	return fmt.Sprintf("https://iframe.mediadelivery.net/embed/%s/%s", c.libraryID, videoID), nil
}

// DirectPlayURL returns a standalone page URL that opens Bunny's player.
//
// Use this when you want to share a direct watch link with a user rather
// than embedding it in a page.
//
// Only requires LibraryID and the video ID. CDNHostname is not needed.
//
//	https://video.bunnycdn.com/play/123/video-guid
func (c *Client) DirectPlayURL(videoID string) (string, error) {
	if strings.TrimSpace(videoID) == "" {
		return "", ErrVideoIDRequired
	}
	return fmt.Sprintf("https://video.bunnycdn.com/play/%s/%s", c.libraryID, videoID), nil
}

// HLSPlaylistURL returns the adaptive bitrate stream manifest URL (.m3u8).
//
// This is the right URL when you want to use your own video player (hls.js,
// video.js, AVPlayer on iOS, ExoPlayer on Android). The player reads the
// manifest and automatically switches quality based on the viewer's bandwidth.
// This is NOT a direct video file — it is a manifest that points to video
// chunks on Bunny's CDN.
//
// Requires CDNHostname to be set in Config.
//
//	https://vz-abc123.b-cdn.net/video-guid/playlist.m3u8
func (c *Client) HLSPlaylistURL(videoID string) (string, error) {
	if strings.TrimSpace(videoID) == "" {
		return "", ErrVideoIDRequired
	}
	base, err := c.cdnBase(videoID)
	if err != nil {
		return "", err
	}
	return base + "/playlist.m3u8", nil
}

// ThumbnailURL returns the static preview image URL for a video.
//
// Requires CDNHostname to be set in Config.
//
//	https://vz-abc123.b-cdn.net/video-guid/thumbnail.jpg
func (c *Client) ThumbnailURL(videoID string) (string, error) {
	if strings.TrimSpace(videoID) == "" {
		return "", ErrVideoIDRequired
	}
	base, err := c.cdnBase(videoID)
	if err != nil {
		return "", err
	}
	return base + "/thumbnail.jpg", nil
}

// PreviewAnimationURL returns the animated WebP preview URL for a video.
//
// This is a short looping animation useful for hover previews in a video
// grid — similar to how YouTube and Netflix animate thumbnails on hover.
//
// Requires CDNHostname to be set in Config.
//
//	https://vz-abc123.b-cdn.net/video-guid/preview.webp
func (c *Client) PreviewAnimationURL(videoID string) (string, error) {
	if strings.TrimSpace(videoID) == "" {
		return "", ErrVideoIDRequired
	}
	base, err := c.cdnBase(videoID)
	if err != nil {
		return "", err
	}
	return base + "/preview.webp", nil
}

// MP4URL returns a direct MP4 download URL at the specified resolution.
//
// Use this when you need a plain downloadable video file — for example,
// for offline viewing, legacy device support, or download links.
//
// IMPORTANT: Requires MP4 Fallback to be enabled in your Bunny library
// settings before uploading the video. Videos uploaded before enabling
// MP4 Fallback will not have MP4 files available.
//
// Requires CDNHostname to be set in Config.
//
//	https://vz-abc123.b-cdn.net/video-guid/play_720p.mp4
func (c *Client) MP4URL(videoID string, r resolution) (string, error) {
	if strings.TrimSpace(videoID) == "" {
		return "", ErrVideoIDRequired
	}
	if r == "" {
		return "", ErrResolutionRequired
	}
	base, err := c.cdnBase(videoID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/play_%s.mp4", base, r), nil
}
