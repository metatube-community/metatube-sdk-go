//go:build vips

package imageutil

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"io"

	"github.com/metatube-community/metatube-sdk-go/imageutil/libvips"
	_ "github.com/metatube-community/metatube-sdk-go/imageutil/libvips"
)

func Decode(r io.Reader) (image.Image, string, error) {
	var buf bytes.Buffer
	tee := io.TeeReader(r, &buf)
	m, _, err := image.Decode(tee)
	var jpegErr jpeg.UnsupportedError
	if err != nil && (errors.Is(err, image.ErrFormat) || errors.As(err, &jpegErr)) {
		// Retry with libvips.
		m, err = libvips.Decode(&buf)
	}
	// m, err := libvips.Decode(r)
	return m, "vips", err
}

func DecodeConfig(r io.Reader) (image.Config, string, error) {
	c, err := libvips.DecodeConfig(r)
	return c, "vips", err
}
