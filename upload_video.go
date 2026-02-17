package bunnystream

import (
	"context"
	"io"
	"net/http"
	"strings"
)

type Resolution string
type OutputCodex string

const (
	Res240p  Resolution = "240p"
	Res360p  Resolution = "360p"
	Res480p  Resolution = "480p"
	Res720p  Resolution = "720p"
	Res1080p Resolution = "1080p"
	Res1440p Resolution = "1440p"
	Res2160p Resolution = "2160p"
)

const (
	Codec_x264 OutputCodex = "x264"
	Codec_vp9  OutputCodex = "vp9"
)

type UploadVideoOptions struct {
	jitEnabled          *bool
	enabledResolution   []string
	enabledOutputCodecs []string
	transcribeEnabled   *bool
	transcribeLanguage  []string
	sourceLanguage      string
	generateTitle       *bool
	genereateDesc       *bool
	generateChapter     *bool
	generateMoments     *bool
}

type UploadVideoOption func(*UploadVideoOptions)

func JITEnabled(v bool) UploadVideoOption {
	return func(o *UploadVideoOptions) {
		o.jitEnabled = &v
	}
}

func EnabledResolutions(resolutions ...Resolution) UploadVideoOption {
	return func(o *UploadVideoOptions) {
		for _, res := range resolutions {
			o.enabledResolution = append(o.enabledResolution, string(res))
		}
	}
}

func EnabledOutputCodexs(codexes ...OutputCodex) UploadVideoOption {
	return func(o *UploadVideoOptions) {
		for _, codex := range codexes {
			o.enabledOutputCodecs = append(o.enabledOutputCodecs, string(codex))
		}
	}
}

func TranscribeEnabled(v bool) UploadVideoOption {
	return func(o *UploadVideoOptions) {
		o.transcribeEnabled = &v
	}
}

func TranscribeLanguages(languages ...string) UploadVideoOption {
	return func(o *UploadVideoOptions) {
		o.transcribeLanguage = append(o.transcribeLanguage, languages...)

	}
}

func SourceLanguage(language string) UploadVideoOption {
	return func(o *UploadVideoOptions) {
		o.sourceLanguage = language
	}
}

func GenerateTitle(v bool) UploadVideoOption {
	return func(o *UploadVideoOptions) {
		o.generateTitle = &v
	}
}

func GenerateDescription(v bool) UploadVideoOption {
	return func(o *UploadVideoOptions) {
		o.genereateDesc = &v
	}
}

func GenerateChapters(v bool) UploadVideoOption {
	return func(o *UploadVideoOptions) {
		o.generateChapter = &v
	}
}

func GenerateMoments(v bool) UploadVideoOption {
	return func(o *UploadVideoOptions) {
		o.generateMoments = &v
	}
}

func FromVideoOption(v UploadVideoOptions) UploadVideoOption {
	return func(o *UploadVideoOptions) {
		*o = v
	}
}

func (c *Client) UploadVideo(ctx context.Context, videoId string, videoFile io.Reader, opts ...UploadVideoOption) (*Response, error) {
	if strings.TrimSpace(videoId) == "" {
		return nil, ErrVideoIDRequired
	}

	uri := c.buildURL("/library/%v/videos/%v", c.libraryID, videoId)

	options := &UploadVideoOptions{}
	for _, opt := range opts {
		opt(options)
	}

	req, err := c.request(ctx, http.MethodPut, uri, videoFile, "application/octet-stream")
	if err != nil {
		return nil, err
	}

	// query := req.URL.Query()

	// if options.jitEnabled != nil {
	// 	query.Set("jitEnabled", strconv.FormatBool(*options.jitEnabled))
	// }
	//
	// if options.enabledResolution != nil {
	// 	query.Set("enabledResolutions", strings.Join(options.enabledResolution, ","))
	// }
	//
	// if options.enabledOutputCodecs != nil {
	// 	query.Set("enabledOutputCodecs", strings.Join(options.enabledOutputCodecs, ","))
	// }
	//
	// if options.transcribeEnabled != nil {
	// 	query.Set("transcribeEnabled", strconv.FormatBool(*options.transcribeEnabled))
	// }
	//
	// if options.transcribeLanguage != nil {
	// 	query.Set("transcribeLanguages", strings.Join(options.transcribeLanguage, ","))
	// }
	//
	// if options.sourceLanguage != "" {
	// 	query.Set("sourceLanguage", options.sourceLanguage)
	// }
	//
	// if options.generateTitle != nil {
	// 	query.Set("generateTitle", strconv.FormatBool(*options.generateTitle))
	// }

	// if options.genereateDesc != nil {
	// 	query.Set("generateDescription", strconv.FormatBool(*options.genereateDesc))
	// }
	//
	// if options.generateChapter != nil {
	// 	query.Set("generateChapters", strconv.FormatBool(*options.generateChapter))
	// }
	//
	// if options.generateMoments != nil {
	// 	query.Set("generateMoments", strconv.FormatBool(*options.generateMoments))
	// }

	buildQuery(req).
		setBool("jitEnabled", options.jitEnabled).
		setStrings("enabledResolutions", options.enabledResolution).
		setStrings("enabledOutputCodecs", options.enabledOutputCodecs).
		setBool("transcribeEnabled", options.transcribeEnabled).
		setStrings("transcribeLanguages", options.transcribeLanguage).
		setString("sourceLanguage", options.sourceLanguage).
		setBool("generateTitle", options.generateTitle).
		setBool("generateDescription", options.genereateDesc).
		setBool("generateChapters", options.generateChapter).
		setBool("generateMoments", options.generateMoments).
		apply()

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
