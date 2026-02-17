package bunnystream

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// mustNewRequest creates a GET request for use in tests.
// It calls t.Fatal if the request cannot be created — this should never happen
// in practice since the URL is a hard-coded constant.
func mustNewRequest(t *testing.T) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, "https://video.bunnycdn.com/library/123/videos", nil)
	if err != nil {
		t.Fatalf("mustNewRequest: %v", err)
	}
	return req
}

// mustNewRequestWithQuery creates a GET request with pre-existing query params.
func mustNewRequestWithQuery(t *testing.T, existing url.Values) *http.Request {
	t.Helper()
	req := mustNewRequest(t)
	req.URL.RawQuery = existing.Encode()
	return req
}

// boolPtr returns a pointer to a bool. Avoids the &true syntax which Go doesn't allow.
func boolPtr(v bool) *bool { return &v }

// -----------------------------------------------------------------------------
// buildQuery
// -----------------------------------------------------------------------------

func TestBuildQuery_ReturnsNonNilBuilder(t *testing.T) {
	req := mustNewRequest(t)
	if buildQuery(req) == nil {
		t.Fatal("buildQuery returned nil")
	}
}

func TestBuildQuery_DoesNotMutateRequestBeforeapply(t *testing.T) {
	req := mustNewRequest(t)
	originalQuery := req.URL.RawQuery

	// Build params but intentionally do NOT call apply.
	buildQuery(req).
		setBool("jitEnabled", boolPtr(true)).
		setString("sourceLanguage", "en").
		setStrings("resolutions", []string{"720p"})

	if req.URL.RawQuery != originalQuery {
		t.Errorf("request URL was mutated before apply: got %q, want %q",
			req.URL.RawQuery, originalQuery)
	}
}

// -----------------------------------------------------------------------------
// setBool
// -----------------------------------------------------------------------------

func TestSetBool(t *testing.T) {
	tests := []struct {
		name      string
		input     *bool
		wantset   bool
		wantValue string
	}{
		{
			name:      "true sets param to 'true'",
			input:     boolPtr(true),
			wantset:   true,
			wantValue: "true",
		},
		{
			name:      "false sets param to 'false'",
			input:     boolPtr(false),
			wantset:   true,
			wantValue: "false",
		},
		{
			name:    "nil pointer omits param entirely",
			input:   nil,
			wantset: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mustNewRequest(t)
			buildQuery(req).setBool("flag", tt.input).apply()

			got := req.URL.Query().Get("flag")
			if tt.wantset && got != tt.wantValue {
				t.Errorf("flag = %q, want %q", got, tt.wantValue)
			}
			if !tt.wantset && got != "" {
				t.Errorf("expected flag to be absent, got %q", got)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// setString
// -----------------------------------------------------------------------------

func TestSetString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantset bool
	}{
		{
			name:    "non-empty string sets param",
			input:   "en",
			wantset: true,
		},
		{
			name:    "empty string omits param",
			input:   "",
			wantset: false,
		},
		{
			name:    "whitespace-only omits param",
			input:   "   ",
			wantset: false,
		},
		{
			name:    "tab-only omits param",
			input:   "\t",
			wantset: false,
		},
		{
			name:    "string with internal spaces is preserved",
			input:   "en US",
			wantset: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mustNewRequest(t)
			buildQuery(req).setString("lang", tt.input).apply()

			got := req.URL.Query().Get("lang")
			if tt.wantset && got != tt.input {
				t.Errorf("lang = %q, want %q", got, tt.input)
			}
			if !tt.wantset && got != "" {
				t.Errorf("expected lang to be absent, got %q", got)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// setStrings
// -----------------------------------------------------------------------------

func TestSetStrings(t *testing.T) {
	tests := []struct {
		name      string
		input     []string
		wantset   bool
		wantValue string
	}{
		{
			name:      "single value",
			input:     []string{"720p"},
			wantset:   true,
			wantValue: "720p",
		},
		{
			name:      "multiple values joined with comma",
			input:     []string{"480p", "720p", "1080p"},
			wantset:   true,
			wantValue: "480p,720p,1080p",
		},
		{
			name:    "nil slice omits param",
			input:   nil,
			wantset: false,
		},
		{
			name:    "empty slice omits param",
			input:   []string{},
			wantset: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mustNewRequest(t)
			buildQuery(req).setStrings("resolutions", tt.input).apply()

			got := req.URL.Query().Get("resolutions")
			if tt.wantset && got != tt.wantValue {
				t.Errorf("resolutions = %q, want %q", got, tt.wantValue)
			}
			if !tt.wantset && got != "" {
				t.Errorf("expected resolutions to be absent, got %q", got)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// apply
// -----------------------------------------------------------------------------

func TestApply_WritesQueryToRequest(t *testing.T) {
	req := mustNewRequest(t)
	buildQuery(req).setString("key", "value").apply()

	if req.URL.Query().Get("key") != "value" {
		t.Errorf("apply did not write query params to the request URL")
	}
}

func TestApply_PreservesExistingQueryParams(t *testing.T) {
	existing := url.Values{"page": []string{"2"}}
	req := mustNewRequestWithQuery(t, existing)

	buildQuery(req).setString("lang", "en").apply()

	q := req.URL.Query()
	if q.Get("page") != "2" {
		t.Error("apply dropped existing query param 'page'")
	}
	if q.Get("lang") != "en" {
		t.Error("apply did not add new query param 'lang'")
	}
}

func TestApply_OverwritesExistingParamWithSameKey(t *testing.T) {
	existing := url.Values{"lang": []string{"fr"}}
	req := mustNewRequestWithQuery(t, existing)

	buildQuery(req).setString("lang", "en").apply()

	got := req.URL.Query().Get("lang")
	if got != "en" {
		t.Errorf("expected existing param to be overwritten with 'en', got %q", got)
	}
}

func TestApply_IsIdempotent(t *testing.T) {
	req := mustNewRequest(t)
	qb := buildQuery(req).setString("lang", "en")

	qb.apply()
	qb.apply()

	q := req.URL.Query()
	if vals := q["lang"]; len(vals) != 1 {
		t.Errorf("expected exactly one 'lang' param after double apply, got %v", vals)
	}
}

// -----------------------------------------------------------------------------
// Chaining
// -----------------------------------------------------------------------------

func TestChaining_AllsettersTogether(t *testing.T) {
	req := mustNewRequest(t)

	buildQuery(req).
		setBool("jitEnabled", boolPtr(true)).
		setBool("transcribeEnabled", boolPtr(false)).
		setString("sourceLanguage", "en").
		setStrings("enabledResolutions", []string{"720p", "1080p"}).
		setStrings("enabledOutputCodecs", []string{"x264", "vp9"}).
		apply()

	q := req.URL.Query()

	cases := []struct{ key, want string }{
		{"jitEnabled", "true"},
		{"transcribeEnabled", "false"},
		{"sourceLanguage", "en"},
		{"enabledResolutions", "720p,1080p"},
		{"enabledOutputCodecs", "x264,vp9"},
	}
	for _, c := range cases {
		if got := q.Get(c.key); got != c.want {
			t.Errorf("%s = %q, want %q", c.key, got, c.want)
		}
	}
}

func TestChaining_NilAndEmptyValuesAreSkipped(t *testing.T) {
	req := mustNewRequest(t)

	buildQuery(req).
		setBool("nilBool", nil).
		setString("emptyString", "").
		setString("whitespace", "   ").
		setStrings("emptySlice", []string{}).
		setStrings("nilSlice", nil).
		apply()

	if raw := req.URL.RawQuery; raw != "" {
		t.Errorf("expected empty query string when all values are zero/nil, got %q", raw)
	}
}

// -----------------------------------------------------------------------------
// Benchmarks
// -----------------------------------------------------------------------------

// BenchmarkBuildQuery_FullChain measures the overhead of building and applying
// a realistic set of query params — representative of an UploadVideo call.
func BenchmarkBuildQuery_FullChain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodPut, "https://video.bunnycdn.com/library/123/videos/abc", nil)
		buildQuery(req).
			setBool("jitEnabled", boolPtr(true)).
			setStrings("enabledResolutions", []string{"480p", "720p", "1080p"}).
			setStrings("enabledOutputCodecs", []string{"x264"}).
			setBool("transcribeEnabled", boolPtr(true)).
			setStrings("transcribeLanguages", []string{"en", "es"}).
			setString("sourceLanguage", "en").
			setBool("generateTitle", boolPtr(true)).
			setBool("generateDescription", boolPtr(false)).
			setBool("generateChapters", boolPtr(false)).
			setBool("generateMoments", boolPtr(true)).
			apply()
	}
}

// -----------------------------------------------------------------------------
// Examples — rendered as runnable code on pkg.go.dev
// -----------------------------------------------------------------------------

func ExampleQueryBuilder_basic() {
	req, _ := http.NewRequest(http.MethodPut, "https://video.bunnycdn.com/library/123/videos/abc", nil)

	jit := true
	buildQuery(req).
		setBool("jitEnabled", &jit).
		setString("sourceLanguage", "en").
		setStrings("enabledResolutions", []string{"720p", "1080p"}).
		apply()

	fmt.Println(req.URL.RawQuery)
	// Output: enabledResolutions=720p%2C1080p&jitEnabled=true&sourceLanguage=en
}

func ExampleQueryBuilder_nilAndEmptyValuesAreIgnored() {
	req, _ := http.NewRequest(http.MethodPut, "https://video.bunnycdn.com/library/123/videos/abc", nil)

	buildQuery(req).
		setBool("jitEnabled", nil).      // nil pointer — skipped
		setString("sourceLanguage", ""). // empty string — skipped
		setStrings("resolutions", nil).  // nil slice — skipped
		apply()

	fmt.Println(req.URL.RawQuery)
	// Output:
}
