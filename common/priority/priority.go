package priority

import "sort"

var _ sort.Interface = (*Slice[int, int])(nil)

type Slice[T int | float64, U any] struct {
	priorities []T
	underlying []U
}

func (s *Slice[T, U]) Append(priority T, item U) {
	s.priorities = append(s.priorities, priority)
	s.underlying = append(s.underlying, item)
}

func (s *Slice[T, U]) Sort() *Slice[T, U] {
	sort.Sort(s)
	return s
}

func (s *Slice[T, U]) Stable() *Slice[T, U] {
	sort.Stable(s)
	return s
}

func (s *Slice[T, U]) Reverse() *Slice[T, U] {
	sort.Stable(sort.Reverse(s))
	return s
}

func (s *Slice[_, U]) Underlying() []U {
	return s.underlying
}

func (s *Slice[_, _]) Len() int {
	return len(s.priorities)
}

func (s *Slice[_, _]) Swap(i int, j int) {
	s.priorities[i], s.priorities[j] = s.priorities[j], s.priorities[i]
	s.underlying[i], s.underlying[j] = s.underlying[j], s.underlying[i]
}

func (s *Slice[_, _]) Less(i int, j int) bool {
	// higher priority comes first.
	return s.priorities[i] > s.priorities[j]
}
