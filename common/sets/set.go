package sets

import (
	"encoding/json"
	"iter"
	"slices"

	"github.com/elliotchance/orderedmap/v2"
)

var (
	_ json.Marshaler   = (*OrderedSet[int, any])(nil)
	_ json.Unmarshaler = (*OrderedSet[int, any])(nil)
)

type OrderedSet[K comparable, V any] struct {
	h func(V) K
	m *orderedmap.OrderedMap[K, V]
}

func NewOrderedSet[K comparable, V any](hash func(V) K) *OrderedSet[K, V] {
	return &OrderedSet[K, V]{
		h: hash,
		m: orderedmap.NewOrderedMap[K, V](),
	}
}

func (set *OrderedSet[K, V]) Len() int {
	return set.m.Len()
}

func (set *OrderedSet[K, V]) Add(items ...V) {
	for _, result := range items {
		set.m.Set(set.h(result), result)
	}
}

func (set *OrderedSet[K, V]) Del(items ...V) {
	for _, result := range items {
		set.m.Delete(set.h(result))
	}
}

func (set *OrderedSet[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, v := range set.m.Iterator() {
			if !yield(v) {
				return
			}
		}
	}
}

func (set *OrderedSet[K, V]) Slice() []V {
	return slices.Collect(set.Values())
}

func (set *OrderedSet[K, V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(set.Slice())
}

func (set *OrderedSet[K, V]) UnmarshalJSON(data []byte) error {
	vs := make([]V, 0)
	if err := json.Unmarshal(data, &vs); err != nil {
		return err
	}
	set.Add(vs...)
	return nil
}
