package imageutil

import (
	"errors"
	"image"
	"net/http"
	"net/url"
	"os"
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
	var (
		req  *http.Request
		resp *http.Response
	)
	if req, err = http.NewRequest(http.MethodGet, u, nil); err != nil {
		return
	}
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", errors.New(http.StatusText(resp.StatusCode))
	}
	return image.Decode(resp.Body)
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
