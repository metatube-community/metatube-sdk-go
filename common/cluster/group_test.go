package cluster

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

var _ WeightedLocatable[weightedPoint, float64, float64] = (*weightedPoint)(nil)

// weightedPoint is a simple 2D point that implements Locatable[weightedPoint].
type weightedPoint struct {
	X, Y, W float64
}

func (a weightedPoint) DistanceTo(b weightedPoint) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return math.Hypot(dx, dy)
}

func (a weightedPoint) Weight() float64 {
	return a.W
}

func TestGroupByDistanceAndSort(t *testing.T) {
	points := []weightedPoint{
		{0.10, 0.10, 1.0},
		{0.12, 0.10, 1.0},
		{0.13, 0.12, 2.0},
		{0.90, 0.90, 3.0},
		{0.91, 0.93, 3.0},
		{0.52, 0.90, 5.0},
	}

	threshold := 0.05

	// Group points by distance.
	groups := GroupByDistance[weightedPoint, float64](points, threshold)

	// Assert group count.
	assert.Len(t, groups, 3)

	// Sort groups by size (descending) and verify.
	SortGroupsBySize(groups)
	assert.Len(t, groups[0].Items, 3)
	assert.Len(t, groups[1].Items, 2)
	assert.Len(t, groups[2].Items, 1)

	// Sort groups by total weight (descending) and verify.
	SortGroupsByWeight(groups)
	assert.Len(t, groups[0].Items, 2)
	assert.Len(t, groups[1].Items, 1)
	assert.Len(t, groups[2].Items, 3)
}
