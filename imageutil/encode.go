package imageutil

import (
	"image"
	"image/jpeg"
	"io"
)

func EncodeToJPEG(w io.Writer, m image.Image, quality int) error {
	return jpeg.Encode(w, m, &jpeg.Options{Quality: quality})
}
