package geomath

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRotatePoint(t *testing.T) {
	for _, unit := range []struct {
		x, y   int
		w, h   int
		angle  float64
		x1, y1 int
	}{
		{0, 0, 100, 100, 270, 99, 0},
		{99, 0, 100, 100, 90, 0, 0},
		{0, 0, 100, 200, 270, 199, 0},
		{0, 0, 100, 200, 270, 199, 0},
		{99, 99, 200, 200, 90, 99, 99},
		{400, 300, 400, 300, 90, 300, 0},
		{5, 295, 400, 300, 270, 5, 5},
		{5, 5, 300, 400, 90, 5, 295},
	} {
		x1, y1 := RotatePoint(unit.x, unit.y, unit.w, unit.h, unit.angle)
		assert.Truef(t,
			pointEqual([2]int{unit.x1, unit.y1}, [2]int{x1, y1}, 1.0),
			"expect: (%d±1,%d±1), but got: (%d,%d)", unit.x1, unit.y1, x1, y1)
	}
}

func pointEqual(p1, p2 [2]int, delta float64) bool {
	return math.Abs(float64(p1[0]-p2[0])) <= delta &&
		math.Abs(float64(p1[1]-p2[1])) <= delta
}
