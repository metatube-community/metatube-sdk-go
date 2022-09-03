package imageutil

import (
	"image"

	"golang.org/x/image/draw"
)

// Resize provides a simple interface to resize image.
func Resize(src image.Image, scale float64) image.Image {
	rect := image.Rect(0, 0,
		int(float64(src.Bounds().Dx())*scale),
		int(float64(src.Bounds().Dy())*scale))
	dst := image.NewRGBA(rect)
	sc := draw.ApproxBiLinear /* default interpolator */
	sc.Scale(dst, rect, src, src.Bounds(), draw.Over, nil)
	return dst
}
