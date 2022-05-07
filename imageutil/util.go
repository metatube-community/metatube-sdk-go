package imageutil

import (
	"errors"
	"image"
	"net/http"
	"os"
)

// GetImageByPath gets image from path.
func GetImageByPath(p string) (image.Image, string, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, "", err
	}
	return image.Decode(f)
}

// GetImageByURL gets image from url.
func GetImageByURL(url string) (_ image.Image, _ string, err error) {
	var (
		req  *http.Request
		resp *http.Response
	)
	if req, err = http.NewRequest(http.MethodGet, url, nil); err != nil {
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
