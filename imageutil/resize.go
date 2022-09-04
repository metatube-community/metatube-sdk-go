package imageutil

import (
	"image"

	"golang.org/x/image/draw"
)

// Resize provides a simple interface to resize image.
func Resize(src image.Image, width, height int) image.Image {
	switch {
	case width == 0 && height == 0:
		return src /* not modified */
	case width == 0 && height != 0:
		width = int(float64(height) / float64(src.Bounds().Dy()) * float64(src.Bounds().Dx()))
	case width != 0 && height == 0:
		height = int(float64(width) / float64(src.Bounds().Dx()) * float64(src.Bounds().Dy()))
	}
	rect := image.Rect(0, 0, width, height)
	dst := image.NewRGBA(rect)
	sc := draw.BiLinear /* default interpolator */
	sc.Scale(dst, rect, src, src.Bounds(), draw.Over, nil)
	return dst
}
