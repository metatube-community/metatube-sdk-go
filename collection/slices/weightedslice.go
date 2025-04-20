package slices

import (
	"cmp"
	"sort"
)

var _ sort.Interface = (*WeightedSlice[any, int])(nil)

type WeightedSlice[O any, W cmp.Ordered] struct {
	objects []O
	weights []W
}

func NewWeightedSlice[O any, W cmp.Ordered](objects []O, weights []W) *WeightedSlice[O, W] {
	if len(objects) != len(weights) {
		panic("objects and weights must have the same length")
	}
	return &WeightedSlice[O, W]{objects, weights}
}

func (s *WeightedSlice[O, W]) Len() int {
	return len(s.objects)
}

func (s *WeightedSlice[O, W]) Less(i int, j int) bool {
	// higher-weighted item comes first.
	return s.weights[i] > s.weights[j]
}

func (s *WeightedSlice[O, W]) Swap(i int, j int) {
	s.weights[i], s.weights[j] = s.weights[j], s.weights[i]
	s.objects[i], s.objects[j] = s.objects[j], s.objects[i]
}

func (s *WeightedSlice[O, W]) Append(object O, weight W) {
	s.weights = append(s.weights, weight)
	s.objects = append(s.objects, object)
}

func (s *WeightedSlice[O, W]) Slice() []O {
	return s.objects
}

func (s *WeightedSlice[O, W]) SortFunc(sortFn func(p sort.Interface)) *WeightedSlice[O, W] {
	sortFn(s)
	return s
}
