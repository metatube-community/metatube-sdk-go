package collections

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWeightedSlice(t *testing.T) {
	s := &WeightedSlice[int, int]{}
	s.Append(2, 5)
	s.Append(3, 1)
	s.Append(1, 9)
	s.Append(6, -1)
	s.Append(6, 9)
	s.Append(6, 8)
	s.Append(6, 7)
	s.Append(6, 0)
	s.Append(0, 12)
	assert.Equal(t,
		[]int{-1, 9, 8, 7, 0, 1, 5, 9, 12},
		s.SortFunc(sort.Stable).Underlying())
}
