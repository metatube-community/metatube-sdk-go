package maps

import (
	"encoding/json"
	"iter"

	"github.com/projectbarks/cimap"
)

var (
	_ json.Marshaler   = (*CaseInsensitiveMap[any])(nil)
	_ json.Unmarshaler = (*CaseInsensitiveMap[any])(nil)
)

type CaseInsensitiveMap[T any] struct {
	internalMap *cimap.CaseInsensitiveMap[T]
}

func NewCaseInsensitiveMap[T any]() *CaseInsensitiveMap[T] {
	return &CaseInsensitiveMap[T]{
		internalMap: cimap.New[T](),
	}
}

func NewCaseInsensitiveMapWithCapacity[T any](capacity int) *CaseInsensitiveMap[T] {
	return &CaseInsensitiveMap[T]{
		internalMap: cimap.New[T](capacity),
	}
}

func (m *CaseInsensitiveMap[T]) Copy() *CaseInsensitiveMap[T] {
	m2 := NewCaseInsensitiveMapWithCapacity[T](m.Len())
	for key, value := range m.Iterator() {
		m2.Set(key, value)
	}
	return m2
}

func (m *CaseInsensitiveMap[T]) Has(key string) bool {
	_, exist := m.internalMap.Get(key)
	return exist
}

func (m *CaseInsensitiveMap[T]) Get(key string) (T, bool) {
	return m.internalMap.Get(key)
}

func (m *CaseInsensitiveMap[T]) GetOrDefault(key string, defaultValues ...T) T {
	value, exist := m.internalMap.Get(key)
	if exist {
		return value
	}
	if len(defaultValues) > 0 {
		return defaultValues[0]
	}
	var defaultValue T
	return defaultValue
}

func (m *CaseInsensitiveMap[T]) Set(key string, value T) {
	m.internalMap.Add(key, value)
}

func (m *CaseInsensitiveMap[T]) Delete(key string) {
	m.internalMap.Delete(key)
}

func (m *CaseInsensitiveMap[T]) Len() int {
	return m.internalMap.Len()
}

func (m *CaseInsensitiveMap[T]) Keys() iter.Seq[string] {
	return m.internalMap.Keys()
}

func (m *CaseInsensitiveMap[T]) Values() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, value := range m.Iterator() {
			if !yield(value) {
				return
			}
		}
	}
}

func (m *CaseInsensitiveMap[T]) Iterator() iter.Seq2[string, T] {
	return m.internalMap.Iterator()
}

func (m *CaseInsensitiveMap[T]) MarshalJSON() ([]byte, error) {
	return m.internalMap.MarshalJSON()
}

func (m *CaseInsensitiveMap[T]) UnmarshalJSON(data []byte) error {
	return m.internalMap.UnmarshalJSON(data)
}
