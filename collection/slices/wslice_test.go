package slices

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWeightedSlice(t *testing.T) {
	s := NewWeightedSlice(
		// initialized pairs.
		[]int{9}, []float64{3},
	)
	s.Append(5, 2)
	s.Append(1, 3)
	s.Append(9, 1)
	s.Append(-1, 6)
	s.Append(9, 6)
	s.Append(8, 6)
	s.Append(7, 6)
	s.Append(0, 6)
	s.Append(12, 0)

	exp := []int{-1, 9, 8, 7, 0, 9, 1, 5, 9, 12}
	got := s.SortFunc(sort.Stable).Slice()
	assert.Equal(t, exp, got)
}
