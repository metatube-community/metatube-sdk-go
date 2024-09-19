//go:build vips

package imageutil

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"io"

	"github.com/metatube-community/metatube-sdk-go/imageutil/vips"
	_ "github.com/metatube-community/metatube-sdk-go/imageutil/vips"
)

func Decode(r io.Reader) (m image.Image, _ string, err error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return
	}
	var jpegErr jpeg.UnsupportedError
	m, _, err = image.Decode(bytes.NewBuffer(buf))
	if err != nil && (errors.Is(err, image.ErrFormat) || errors.As(err, &jpegErr)) {
		// retry to decode with libvips.
		m, err = vips.Decode(bytes.NewBuffer(buf))
	}
	return m, "vips", err
}

func DecodeConfig(r io.Reader) (image.Config, string, error) {
	c, err := vips.DecodeConfig(r)
	return c, "vips", err
}
