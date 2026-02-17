# bunnystream-go

A Go SDK for the [Bunny Stream](https://bunny.net/stream/) API. Handles video object creation, file uploads, URL generation, and signed URL authentication.

## Installation

```bash
go get github.com/ArabindaSigdel/bunnystream-go
```

Requires Go 1.25+.

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    bunnystream "github.com/ArabindaSigdel/bunnystream-go"
)

func main() {
    client, err := bunnystream.NewClient(&bunnystream.Config{
        APIKey:      os.Getenv("BUNNY_API_KEY"),
        LibraryID:   os.Getenv("BUNNY_LIBRARY_ID"),
        CDNHostname: os.Getenv("BUNNY_CDN_HOSTNAME"), // e.g. "vz-abc123.b-cdn.net"
    })
    if err != nil {
        log.Fatal(err)
    }

    // 1. Create a video object
    resp, err := client.CreateVideoObject(context.Background(), "My Video")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Created:", string(resp.Body))

    // 2. Upload the video file
    f, _ := os.Open("video.mp4")
    defer f.Close()

    _, err = client.UploadVideo(context.Background(), "your-video-id", f)
    if err != nil {
        log.Fatal(err)
    }

    // 3. Get a playback URL
    embedURL, _ := client.EmbedURL("your-video-id")
    fmt.Println("Watch at:", embedURL)
}
```

## Configuration

```go
client, err := bunnystream.NewClient(&bunnystream.Config{
    // Required
    APIKey:    "your-api-key",   // Stream library API key
    LibraryID: "123456",         // Stream library ID

    // Required for HLS, thumbnail, preview, and MP4 URLs
    CDNHostname: "vz-abc123.b-cdn.net",

    // Required only when Embed View Token Authentication is enabled
    EmbedTokenKey: "your-embed-token-key",

    // Required only when CDN Token Authentication is enabled on your pull zone
    CDNTokenKey: "your-cdn-token-key",

    // Optional — these have sensible defaults
    Timeout:    30 * time.Second, // default: 60s
    MaxRetries: 3,                // default: 3
    UserAgent:  "my-app/1.0",    // default: "bunnystream-go/0.1.0"
})
```

> **Security:** Never hardcode `APIKey`, `EmbedTokenKey`, or `CDNTokenKey` in your source code. Load them from environment variables or a secrets manager. These values must only ever be used server-side.

## Usage

### Create a Video Object

Before uploading, you need to create a video entry in your library:

```go
resp, err := client.CreateVideoObject(ctx, "My Video",
    bunnystream.WithCollectionID("collection-uuid"),
    bunnystream.WithThumbnailTime("30"),
)
```

### Upload a Video

```go
f, err := os.Open("video.mp4")
if err != nil {
    log.Fatal(err)
}
defer f.Close()

resp, err := client.UploadVideo(ctx, "video-id", f,
    bunnystream.JITEnabled(true),
    bunnystream.EnabledResolutions(bunnystream.Res720p, bunnystream.Res1080p),
    bunnystream.EnabledOutputCodexs(bunnystream.Codec_x264),
    bunnystream.TranscribeEnabled(true),
    bunnystream.TranscribeLanguages("en", "es"),
    bunnystream.GenerateTitle(true),
    bunnystream.GenerateDescription(true),
)
```

### Playback URLs

```go
// Iframe embed — works without CDNHostname
embedURL, _      := client.EmbedURL("video-id")
directURL, _     := client.DirectPlayURL("video-id")

// CDN-backed URLs — require CDNHostname in Config
hlsURL, _        := client.HLSPlaylistURL("video-id")
thumbnailURL, _  := client.ThumbnailURL("video-id")
previewURL, _    := client.PreviewAnimationURL("video-id")
mp4URL, _        := client.MP4URL("video-id", bunnystream.Res1080p)
```

### Signed URLs

Use signed URLs when token authentication is enabled on your library or pull zone.

```go
// Signed iframe embed (requires EmbedTokenKey in Config)
signedEmbed, err := client.SignedEmbedURL("video-id", 2*time.Hour)

// Signed HLS stream — uses directory token so .ts chunks work too (requires CDNTokenKey)
signedHLS, err := client.SignedHLSURL("video-id", 2*time.Hour)

// Signed MP4 download (requires CDNTokenKey)
signedMP4, err := client.SignedMP4URL("video-id", bunnystream.Res720p, 24*time.Hour)

// Optional: restrict by IP or country
signedHLS, err = client.SignedHLSURL("video-id", 2*time.Hour,
    bunnystream.WithUserIP("1.2.3.4"),
    bunnystream.WithCountriesAllowed("US,GB"),
)

signedMP4, err = client.SignedMP4URL("video-id", bunnystream.Res720p, 24*time.Hour,
    bunnystream.WithCountriesBlocked("CN,RU"),
)
```

## Error Handling

All errors can be checked with `errors.Is`:

```go
resp, err := client.UploadVideo(ctx, videoID, file)
if err != nil {
    switch {
    case errors.Is(err, bunnystream.ErrUnauthorized):
        // wrong API key
    case errors.Is(err, bunnystream.ErrForbidden):
        // token auth failed, geo-block, or insufficient permissions
    case errors.Is(err, bunnystream.ErrVideoNotFound):
        // video ID doesn't exist in this library
    case errors.Is(err, bunnystream.ErrRateLimited):
        // back off and retry
    default:
        // check for structured API errors
        var apiErr *bunnystream.APIError
        if errors.As(err, &apiErr) {
            fmt.Printf("API error %d: %s\n", apiErr.StatusCode, apiErr.Message)
        }
    }
}
```

### Sentinel Errors

| Error | When it's returned |
|---|---|
| `ErrAPIKeyRequired` | `APIKey` missing from Config |
| `ErrLibraryIDRequired` | `LibraryID` missing from Config |
| `ErrVideoIDRequired` | empty video ID passed to any method |
| `ErrTitleRequired` | empty title passed to `CreateVideoObject` |
| `ErrResolutionRequired` | empty resolution passed to `MP4URL` / `SignedMP4URL` |
| `ErrCDNHostnameRequired` | CDN URL method called without `CDNHostname` in Config |
| `ErrEmbedTokenKeyRequired` | `SignedEmbedURL` called without `EmbedTokenKey` in Config |
| `ErrCDNTokenKeyRequired` | signed CDN URL method called without `CDNTokenKey` in Config |
| `ErrUnauthorized` | API returned 401 |
| `ErrForbidden` | API returned 403 |
| `ErrVideoNotFound` | API returned 404 |
| `ErrRateLimited` | API returned 429 |
| `ErrInternalServer` | API returned 500 |
| `ErrServiceUnavailable` | API returned 503 |

## Known Limitations

- API responses are returned as raw `[]byte` in `Response.Body`. Typed response structs (e.g. `Video`, `Collection`) are planned for v0.2.
- No automatic retry logic yet. `MaxRetries` is accepted in Config but not yet implemented.

## License

MIT — see [LICENSE](LICENSE).
