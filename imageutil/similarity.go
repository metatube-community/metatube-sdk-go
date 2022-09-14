package imageutil

import (
	"image"

	"github.com/vitali-fedulov/images4"
)

// Similar is a shortcut of images4.Similar.
func Similar(imgA, imgB image.Image) bool {
	// Icons are compact image representations (image "hashes").
	// Name "hash" is not used intentionally.
	iconA := images4.Icon(imgA)
	iconB := images4.Icon(imgB)

	// Comparison.
	// Images are not used directly. Icons are used instead,
	// because they have tiny memory footprint and fast to compare.
	return images4.Similar(iconA, iconB)
}
