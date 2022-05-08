package imageutil

import (
	// Register Formats
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

func min[T int | float64](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func max[T int | float64](a, b T) T {
	if a > b {
		return a
	}
	return b
}
