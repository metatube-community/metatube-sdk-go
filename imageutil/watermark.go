package imageutil

import (
	"image"
	"image/draw"
)

func Watermark(src image.Image, wmk image.Image, pt image.Point) image.Image {
	dst := image.NewNRGBA(image.Rect(0, 0,
		src.Bounds().Dx(), src.Bounds().Dy()))
	draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Src)
	draw.Draw(dst, dst.Bounds(), wmk, wmk.Bounds().Min.Add(pt), draw.Over)
	return dst
}
