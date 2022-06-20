package imageutil

import (
	"image"
)

func CropImage(img image.Image, rect image.Rectangle) image.Image {
	return img.(interface {
		SubImage(image.Rectangle) image.Image
	}).SubImage(rect)
}

const (
	minRatio = 1e-2
	maxRatio = 1e2
)

func CropImagePosition(img image.Image, ratio float64, pos float64) image.Image {
	if ratio < minRatio || ratio > maxRatio {
		return img // no cropping
	}
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	var (
		w, h = width, height
		x, y = 0, 0 // default
	)
	if w = int(float64(height) * ratio); w < width {
		x = max(min(int(float64(width)*pos)-int(float64(w)/2), width-w), 0)
	} else if h = int(float64(width) / ratio); h < height {
		y = max(min(int(float64(height)*pos)-int(float64(h)/2), height-h), 0)
	}
	return CropImage(img,
		image.Rect(0, 0, w, h).
			Add(image.Pt(x, y)).Add(img.Bounds().Min))
}
