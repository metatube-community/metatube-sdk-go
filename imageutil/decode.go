package imageutil

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"io"

	"github.com/gen2brain/jpegli"
)

// Ref: https://github.com/gen2brain/jpegli/issues/4
func init() {
	defer func() {
		recover()
	}()
	_, _ = jpegli.Decode(&bytes.Buffer{})
}

func Decode(r io.Reader) (image.Image, string, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, "", err
	}
	var jpegErr jpeg.UnsupportedError
	m, f, err := image.Decode(bytes.NewBuffer(buf))
	if err != nil && errors.As(err, &jpegErr) {
		// Fallback to decode with jpegli.
		m, err = jpegli.Decode(bytes.NewBuffer(buf))
	}
	return m, f, err
}
