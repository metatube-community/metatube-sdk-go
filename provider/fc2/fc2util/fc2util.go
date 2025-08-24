package fc2util

import (
	"bytes"
	_ "embed"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var fc2pattern = regexp.MustCompile(`^(?i)(?:FC2(?:[-_\s]?PPV)?[-_\s]?)?(\d+)$`)

func ParseNumber(id string) string {
	ss := fc2pattern.FindStringSubmatch(id)
	if len(ss) != 2 {
		return ""
	}
	return ss[1]
}

//go:embed fc2-no-image.png
var fc2NoImage []byte

func FetchImage(url string) (*http.Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if !strings.HasSuffix(url, ".jpg") {
		// skip if not requesting an image.
		return resp, nil
	}
	if resp.StatusCode == http.StatusOK ||
		strings.HasPrefix(resp.Header.Get("Content-Type"), "image/") {
		return resp, nil
	}
	defer resp.Body.Close()

	// return no-image content.
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(fc2NoImage)),
	}, nil
}
