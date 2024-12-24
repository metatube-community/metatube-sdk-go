package imageutil

import (
	"errors"
	"image"
	"image/jpeg"
	"io"

	"github.com/docker/go-units"
	"github.com/gen2brain/jpegli"

	"github.com/metatube-community/metatube-sdk-go/common/bufferpool"
)

var _pool = bufferpool.New(256 * units.KiB)

func Decode(r io.Reader) (image.Image, string, error) {
	buf := _pool.Get()
	defer _pool.Put(buf)
	var jpegErr jpeg.UnsupportedError
	m, f, err := image.Decode(io.TeeReader(r, buf))
	if err != nil && errors.As(err, &jpegErr) {
		// Fallback to decode with jpegli.
		m, err = jpegli.Decode(io.MultiReader(buf, r))
	}
	return m, f, err
}
