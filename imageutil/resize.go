package imageutil

import (
	"image"

	"github.com/disintegration/imaging"
)

// Resize provides a simple interface to resize image.
func Resize(src image.Image, width, height int) image.Image {
	switch {
	case width == 0 && height == 0:
		return src /* not modified */
	case width == 0:
		width = int(float64(height) / float64(src.Bounds().Dy()) * float64(src.Bounds().Dx()))
	case height == 0:
		height = int(float64(width) / float64(src.Bounds().Dx()) * float64(src.Bounds().Dy()))
	}
	return imaging.Resize(src, width, height, imaging.Lanczos)
}
