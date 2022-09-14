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
	return similar(iconA, iconB)
}

// Similarity parameters.
const (
	iconSize  = 11
	colorDiff = 50.0
	euclCoeff = 0.2
	chanCoeff = 2.0
)

// Similarity thresholds.
const (
	thY    = colorDiff * colorDiff * euclCoeff * iconSize * iconSize
	thCbCr = colorDiff * colorDiff * euclCoeff * (iconSize + 1) * (iconSize + 1) * chanCoeff
	thProp = 0.05
)

func similar(iconA, iconB images4.IconT) bool {
	if !propSimilar(iconA, iconB) {
		return false
	}
	if !eucSimilar(iconA, iconB) {
		return false
	}
	return true
}

func propSimilar(iconA, iconB images4.IconT) bool {
	return images4.PropMetric(iconA, iconB) < thProp
}

func eucSimilar(iconA, iconB images4.IconT) bool {
	m1, m2, m3 := images4.EucMetric(iconA, iconB)
	return m1 < thY && // Luma as most sensitive.
		m2 < thCbCr &&
		m3 < thCbCr
}
