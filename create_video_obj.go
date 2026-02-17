package bunnystream

import (
	"context"
	"net/http"
	"strings"
)

type videoOptions struct {
	CollectionID  string
	ThumbnailTime string
}

type VideoOption func(*videoOptions)

func WithCollectionID(id string) VideoOption {
	return func(o *videoOptions) {
		o.CollectionID = id
	}
}

func WithThumbnailTime(time string) VideoOption {
	return func(o *videoOptions) {
		o.ThumbnailTime = time
	}
}

// CreateVideoObject initializes a new video entry within a specific library collection.
// It validates that a title is provided and sends a POST request to the video stream API.
//
// Parameters:
//   - title: The display name of the video (Required).
//   - collectionId: The UUID of the collection to which the video belongs (Optional).
//   - thumbnailTime: The timestamp (in seconds/format) to capture the preview image (Optional).
//
// Returns a Response pointer containing the server's metadata or an Error
// if the title is empty.
func (c *Client) CreateVideoObject(ctx context.Context, title string, opts ...VideoOption) (*Response, error) {
	url := c.buildURL("/library/%v/videos", c.libraryID)

	body := make(map[string]string, 1)

	if strings.TrimSpace(title) == "" {
		return nil, ErrTitileRequired
	}
	body["title"] = title

	options := &videoOptions{}

	for _, opt := range opts {
		opt(options)
	}

	if options.CollectionID != "" {
		body["collectionId"] = options.CollectionID
	}

	if options.ThumbnailTime != "" {
		body["thumbnailTime"] = options.ThumbnailTime
	}

	bodyBuf, err := c.encodeJSON(body)
	if err != nil {
		return nil, err
	}

	req, err := c.request(ctx, http.MethodPost, url, bodyBuf, "application/json")
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
