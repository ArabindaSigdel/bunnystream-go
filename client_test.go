package bunnystream

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// testServer creates a fake HTTP server that always returns the given status
// code and body. Returns a client configured to talk to it.
func testServer(t *testing.T, statusCode int, body string) (*Client, *httptest.Server) {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		if body != "" {
			w.Write([]byte(body))
		}
	}))

	cfg := &Config{
		APIKey:     "test-key",
		LibraryID:  "123",
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
	}
	client, err := NewClient(cfg)
	if err != nil {
		srv.Close()
		t.Fatalf("failed to create test client: %v", err)
	}

	return client, srv
}

// inspectServer creates a fake HTTP server that calls the given inspect
// function with each incoming request, so you can assert on what was sent.
func inspectServer(t *testing.T, inspect func(*http.Request), statusCode int) (*Client, *httptest.Server) {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inspect(r)
		w.WriteHeader(statusCode)
	}))

	cfg := &Config{
		APIKey:     "test-key",
		LibraryID:  "123",
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
	}
	client, err := NewClient(cfg)
	if err != nil {
		srv.Close()
		t.Fatalf("failed to create test client: %v", err)
	}

	return client, srv
}

// -----------------------------------------------------------------------------
// checkResponseError — status code mapping
// -----------------------------------------------------------------------------

func TestCheckResponseError_200_NoError(t *testing.T) {
	c, srv := testServer(t, http.StatusOK, `{}`)
	defer srv.Close()

	_, err := c.CreateVideoObject(context.Background(), "My Video")
	if err != nil {
		t.Errorf("expected no error for 200, got %v", err)
	}
}

func TestCheckResponseError_201_NoError(t *testing.T) {
	c, srv := testServer(t, http.StatusCreated, `{}`)
	defer srv.Close()

	_, err := c.CreateVideoObject(context.Background(), "My Video")
	if err != nil {
		t.Errorf("expected no error for 201, got %v", err)
	}
}

func TestCheckResponseError_401_ErrUnauthorized(t *testing.T) {
	c, srv := testServer(t, http.StatusUnauthorized, "")
	defer srv.Close()

	_, err := c.CreateVideoObject(context.Background(), "My Video")
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized for 401, got %v", err)
	}
}

func TestCheckResponseError_403_ErrForbidden(t *testing.T) {
	c, srv := testServer(t, http.StatusForbidden, "")
	defer srv.Close()

	_, err := c.CreateVideoObject(context.Background(), "My Video")
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden for 403, got %v", err)
	}
}

func TestCheckResponseError_404_ErrVideoNotFound(t *testing.T) {
	c, srv := testServer(t, http.StatusNotFound, "")
	defer srv.Close()

	_, err := c.CreateVideoObject(context.Background(), "My Video")
	if !errors.Is(err, ErrVideoNotFound) {
		t.Errorf("expected ErrVideoNotFound for 404, got %v", err)
	}
}

func TestCheckResponseError_429_ErrRateLimited(t *testing.T) {
	c, srv := testServer(t, http.StatusTooManyRequests, "")
	defer srv.Close()

	_, err := c.CreateVideoObject(context.Background(), "My Video")
	if !errors.Is(err, ErrRateLimited) {
		t.Errorf("expected ErrRateLimited for 429, got %v", err)
	}
}

func TestCheckResponseError_500_ErrInternalServer(t *testing.T) {
	c, srv := testServer(t, http.StatusInternalServerError, "")
	defer srv.Close()

	_, err := c.CreateVideoObject(context.Background(), "My Video")
	if !errors.Is(err, ErrInternalServer) {
		t.Errorf("expected ErrInternalServer for 500, got %v", err)
	}
}

func TestCheckResponseError_503_ErrServiceUnavailable(t *testing.T) {
	c, srv := testServer(t, http.StatusServiceUnavailable, "")
	defer srv.Close()

	_, err := c.CreateVideoObject(context.Background(), "My Video")
	if !errors.Is(err, ErrServiceUnavailable) {
		t.Errorf("expected ErrServiceUnavailable for 503, got %v", err)
	}
}

func TestCheckResponseError_400_APIError(t *testing.T) {
	c, srv := testServer(t, http.StatusBadRequest, `invalid input`)
	defer srv.Close()

	_, err := c.CreateVideoObject(context.Background(), "My Video")

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError for 400, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("APIError.StatusCode = %d, want %d", apiErr.StatusCode, http.StatusBadRequest)
	}
}

func TestCheckResponseError_UnknownStatus_APIError(t *testing.T) {
	c, srv := testServer(t, 418, `i'm a teapot`)
	defer srv.Close()

	_, err := c.CreateVideoObject(context.Background(), "My Video")

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError for unhandled status, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 418 {
		t.Errorf("APIError.StatusCode = %d, want 418", apiErr.StatusCode)
	}
}

// -----------------------------------------------------------------------------
// Request construction — CreateVideoObject
// -----------------------------------------------------------------------------

func TestCreateVideoObject_SendsPOST(t *testing.T) {
	var gotMethod string
	c, srv := inspectServer(t, func(r *http.Request) {
		gotMethod = r.Method
	}, http.StatusOK)
	defer srv.Close()

	c.CreateVideoObject(context.Background(), "My Video")

	if gotMethod != http.MethodPost {
		t.Errorf("expected POST, got %q", gotMethod)
	}
}

func TestCreateVideoObject_SendsCorrectPath(t *testing.T) {
	var gotPath string
	c, srv := inspectServer(t, func(r *http.Request) {
		gotPath = r.URL.Path
	}, http.StatusOK)
	defer srv.Close()

	c.CreateVideoObject(context.Background(), "My Video")

	want := "/library/123/videos"
	if gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}

func TestCreateVideoObject_SendsAccessKeyHeader(t *testing.T) {
	var gotKey string
	c, srv := inspectServer(t, func(r *http.Request) {
		gotKey = r.Header.Get("AccessKey")
	}, http.StatusOK)
	defer srv.Close()

	c.CreateVideoObject(context.Background(), "My Video")

	if gotKey != "test-key" {
		t.Errorf("AccessKey header = %q, want %q", gotKey, "test-key")
	}
}

func TestCreateVideoObject_SendsJSONContentType(t *testing.T) {
	var gotCT string
	c, srv := inspectServer(t, func(r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
	}, http.StatusOK)
	defer srv.Close()

	c.CreateVideoObject(context.Background(), "My Video")

	if !strings.HasPrefix(gotCT, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", gotCT)
	}
}

func TestCreateVideoObject_EmptyTitle_ReturnsErrBeforeHTTP(t *testing.T) {
	called := false
	c, srv := inspectServer(t, func(r *http.Request) {
		called = true
	}, http.StatusOK)
	defer srv.Close()

	_, err := c.CreateVideoObject(context.Background(), "")

	if !errors.Is(err, ErrTitleRequired) {
		t.Errorf("expected ErrTitleRequired, got %v", err)
	}
	if called {
		t.Error("HTTP request was made despite empty title — validation should short-circuit")
	}
}

func TestCreateVideoObject_WhitespaceTitleReturnsErrBeforeHTTP(t *testing.T) {
	called := false
	c, srv := inspectServer(t, func(r *http.Request) {
		called = true
	}, http.StatusOK)
	defer srv.Close()

	_, err := c.CreateVideoObject(context.Background(), "   ")

	if !errors.Is(err, ErrTitleRequired) {
		t.Errorf("expected ErrTitleRequired, got %v", err)
	}
	if called {
		t.Error("HTTP request was made despite whitespace title")
	}
}

// -----------------------------------------------------------------------------
// Request construction — UploadVideo
// -----------------------------------------------------------------------------

func TestUploadVideo_SendsPUT(t *testing.T) {
	var gotMethod string
	c, srv := inspectServer(t, func(r *http.Request) {
		gotMethod = r.Method
	}, http.StatusOK)
	defer srv.Close()

	c.UploadVideo(context.Background(), "video-abc", strings.NewReader("fake-data"))

	if gotMethod != http.MethodPut {
		t.Errorf("expected PUT, got %q", gotMethod)
	}
}

func TestUploadVideo_SendsCorrectPath(t *testing.T) {
	var gotPath string
	c, srv := inspectServer(t, func(r *http.Request) {
		gotPath = r.URL.Path
	}, http.StatusOK)
	defer srv.Close()

	c.UploadVideo(context.Background(), "video-abc", strings.NewReader("fake-data"))

	want := "/library/123/videos/video-abc"
	if gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}

func TestUploadVideo_SendsOctetStreamContentType(t *testing.T) {
	var gotCT string
	c, srv := inspectServer(t, func(r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
	}, http.StatusOK)
	defer srv.Close()

	c.UploadVideo(context.Background(), "video-abc", strings.NewReader("fake-data"))

	if gotCT != "application/octet-stream" {
		t.Errorf("Content-Type = %q, want application/octet-stream", gotCT)
	}
}

func TestUploadVideo_EmptyVideoID_ReturnsErrBeforeHTTP(t *testing.T) {
	called := false
	c, srv := inspectServer(t, func(r *http.Request) {
		called = true
	}, http.StatusOK)
	defer srv.Close()

	_, err := c.UploadVideo(context.Background(), "", strings.NewReader("fake-data"))

	if !errors.Is(err, ErrVideoIDRequired) {
		t.Errorf("expected ErrVideoIDRequired, got %v", err)
	}
	if called {
		t.Error("HTTP request was made despite empty videoID — validation should short-circuit")
	}
}

func TestUploadVideo_QueryParamsApplied(t *testing.T) {
	var gotQuery string
	c, srv := inspectServer(t, func(r *http.Request) {
		gotQuery = r.URL.RawQuery
	}, http.StatusOK)
	defer srv.Close()

	c.UploadVideo(
		context.Background(),
		"video-abc",
		strings.NewReader("fake-data"),
		JITEnabled(true),
		EnabledResolutions(Res720p, Res1080p),
		SourceLanguage("en"),
	)

	q, _ := http.NewRequest("GET", "http://x?"+gotQuery, nil)
	params := q.URL.Query()

	if params.Get("jitEnabled") != "true" {
		t.Errorf("jitEnabled = %q, want 'true'", params.Get("jitEnabled"))
	}
	if params.Get("enabledResolutions") != "720p,1080p" {
		t.Errorf("enabledResolutions = %q, want '720p,1080p'", params.Get("enabledResolutions"))
	}
	if params.Get("sourceLanguage") != "en" {
		t.Errorf("sourceLanguage = %q, want 'en'", params.Get("sourceLanguage"))
	}
}

// -----------------------------------------------------------------------------
// APIError
// -----------------------------------------------------------------------------

func TestAPIError_Error_WithMessage(t *testing.T) {
	e := &APIError{StatusCode: 400, Message: "title is required"}
	got := e.Error()

	if !strings.Contains(got, "400") {
		t.Errorf("error string missing status code: %q", got)
	}
	if !strings.Contains(got, "title is required") {
		t.Errorf("error string missing message: %q", got)
	}
}

func TestAPIError_Error_WithoutMessage(t *testing.T) {
	e := &APIError{StatusCode: 418}
	got := e.Error()

	if !strings.Contains(got, "418") {
		t.Errorf("error string missing status code: %q", got)
	}
}
