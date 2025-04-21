package convertor

import (
	"math"
)

// ConvertToCentimeters converts feet and inch to cm.
func ConvertToCentimeters(feet, inches int) int {
	cm := math.Round(float64(feet*12+inches) * 2.54)
	return int(cm)
}
