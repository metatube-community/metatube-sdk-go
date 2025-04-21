package sets

import (
	"encoding/json"
	"iter"
	"slices"

	"github.com/elliotchance/orderedmap/v3"
)

var (
	_ json.Marshaler   = (*OrderedSet[int, any])(nil)
	_ json.Unmarshaler = (*OrderedSet[int, any])(nil)
)

type OrderedSet[K comparable, V any] struct {
	h func(V) K
	m *orderedmap.OrderedMap[K, V]
}

func NewOrderedSet[T comparable]() *OrderedSet[T, T] {
	return &OrderedSet[T, T]{
		h: func(t T) T { return t },
		m: orderedmap.NewOrderedMap[T, T](),
	}
}

func NewOrderedSetWithHash[K comparable, V any](hash func(V) K) *OrderedSet[K, V] {
	return &OrderedSet[K, V]{
		h: hash,
		m: orderedmap.NewOrderedMap[K, V](),
	}
}

func (s *OrderedSet[K, V]) Len() int {
	return s.m.Len()
}

func (s *OrderedSet[K, V]) Add(items ...V) {
	for _, result := range items {
		s.m.Set(s.h(result), result)
	}
}

func (s *OrderedSet[K, V]) Del(items ...V) {
	for _, result := range items {
		s.m.Delete(s.h(result))
	}
}

func (s *OrderedSet[K, V]) Iterator() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, v := range s.m.AllFromFront() {
			if !yield(v) {
				return
			}
		}
	}
}

func (s *OrderedSet[K, V]) AsSlice() []V {
	return slices.Collect(s.Iterator())
}

func (s *OrderedSet[K, V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.AsSlice())
}

func (s *OrderedSet[K, V]) UnmarshalJSON(data []byte) error {
	vs := make([]V, 0)
	if err := json.Unmarshal(data, &vs); err != nil {
		return err
	}
	s.Add(vs...)
	return nil
}
