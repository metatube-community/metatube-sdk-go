//go:build vips

package vips

import (
	"bytes"
	"image"
	"image/png"
	"io"

	"github.com/h2non/bimg"
)

func decode(r io.Reader) (m image.Image, c image.Config, err error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return
	}
	if data, err = bimg.NewImage(data).Process(bimg.Options{
		Type: bimg.PNG,
	}); err != nil {
		return
	}
	if m, err = png.Decode(bytes.NewBuffer(data)); err != nil {
		return
	}
	c = image.Config{
		ColorModel: m.ColorModel(),
		Width:      m.Bounds().Dx(),
		Height:     m.Bounds().Dy(),
	}
	return
}

func Decode(r io.Reader) (image.Image, error) {
	m, _, err := decode(r)
	return m, err
}

func DecodeConfig(r io.Reader) (image.Config, error) {
	_, c, err := decode(r)
	return c, err
}

func init() {
	image.RegisterFormat("vips", "?", Decode, DecodeConfig)
}
