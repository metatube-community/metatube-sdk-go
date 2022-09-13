package imageutil

import (
	"image"

	"github.com/vitali-fedulov/images4"
)

// Similar is a shortcut of images4.Similar.
func Similar(img1, img2 image.Image) bool {
	// Icons are compact image representations (image "hashes").
	// Name "hash" is not used intentionally.
	icon1 := images4.Icon(img1)
	icon2 := images4.Icon(img2)

	// Comparison.
	// Images are not used directly. Icons are used instead,
	// because they have tiny memory footprint and fast to compare.
	return images4.Similar(icon1, icon2)
}
