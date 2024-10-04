package collections

import (
	"sort"

	. "golang.org/x/exp/constraints"
)

var _ sort.Interface = (*WeightedSlice[int, int])(nil)

type WeightedSlice[W Ordered, T any] struct {
	weights []W
	objects []T
}

func (s *WeightedSlice[W, T]) Len() int {
	return len(s.weights)
}

func (s *WeightedSlice[W, T]) Less(i int, j int) bool {
	// higher weighted item comes first.
	return s.weights[i] > s.weights[j]
}

func (s *WeightedSlice[W, T]) Swap(i int, j int) {
	s.weights[i], s.weights[j] = s.weights[j], s.weights[i]
	s.objects[i], s.objects[j] = s.objects[j], s.objects[i]
}

func (s *WeightedSlice[W, T]) Append(weight W, object T) {
	s.weights = append(s.weights, weight)
	s.objects = append(s.objects, object)
}

func (s *WeightedSlice[W, T]) Underlying() []T {
	return s.objects
}

func (s *WeightedSlice[W, T]) SortFunc(fs ...func(p sort.Interface)) *WeightedSlice[W, T] {
	for _, f := range fs {
		f(s)
	}
	return s
}
