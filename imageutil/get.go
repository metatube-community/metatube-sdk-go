package imageutil

import (
	"image"
	"io"
	"net/url"
	"os"

	"github.com/javtube/javtube-sdk-go/common/fetch"
)

// GetImage gets image from url or path.
func GetImage(s string) (image.Image, string, error) {
	if !isValidURL(s) {
		return GetImageByPath(s)
	}
	return GetImageByURL(s)
}

// GetImageByPath gets image from path.
func GetImageByPath(p string) (image.Image, string, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, "", err
	}
	return image.Decode(f)
}

// GetImageByURL gets image from url.
func GetImageByURL(u string) (_ image.Image, _ string, err error) {
	var rc io.ReadCloser
	if rc, err = fetch.Fetch(u); err != nil {
		return
	}
	defer rc.Close()
	return image.Decode(rc)
}

func isValidURL(s string) bool {
	if _, err := url.ParseRequestURI(s); err != nil {
		return false
	}
	if u, err := url.Parse(s); err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	return true
}
