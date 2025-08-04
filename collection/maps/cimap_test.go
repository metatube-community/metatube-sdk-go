package maps

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCaseInsensitiveMap(t *testing.T) {
	m := NewCaseInsensitiveMap[string]()

	// Set with mixed casing
	m.Set("FOO", "bar")
	m.Set("Baz", "qux")
	assert.Equal(t, 2, m.Len())

	// Should be case-insensitive
	val, ok := m.Get("foo")
	assert.True(t, ok)
	assert.Equal(t, "bar", val)

	val, ok = m.Get("baz")
	assert.True(t, ok)
	assert.Equal(t, "qux", val)

	val, ok = m.Get("BAZ")
	assert.True(t, ok)
	assert.Equal(t, "qux", val)

	exist := m.Has("foo")
	assert.True(t, exist)

	exist = m.Has("baz")
	assert.True(t, exist)

	exist = m.Has("BAZ")
	assert.True(t, exist)

	val = m.GetOrDefault("foo", "quux")
	assert.Equal(t, "bar", val)

	val = m.GetOrDefault("baz", "quux")
	assert.Equal(t, "qux", val)

	val = m.GetOrDefault("bar", "quux")
	assert.Equal(t, "quux", val)

	val = m.GetOrDefault("bar")
	assert.Equal(t, "", val)

	keys := slices.Collect(m.Keys())
	slices.Sort(keys)
	assert.Equal(t, []string{"Baz", "FOO"}, keys)

	values := slices.Collect(m.Values())
	slices.Sort(values)
	assert.Equal(t, []string{"bar", "qux"}, values)

	// Delete should also be case-insensitive
	m.Delete("FOO")
	_, ok = m.Get("foo")
	assert.False(t, ok)
	assert.Equal(t, 1, m.Len())

	// Test JSON marshal/unmarshal
	data, err := json.Marshal(m)
	if assert.NoError(t, err) {
		assert.JSONEq(t, `{
			"Baz":"qux"
		}`, string(data))
	}

	copied := m.Copy()
	data2, err := json.Marshal(copied)
	if assert.NoError(t, err) {
		assert.JSONEq(t, `{
			"Baz":"qux"
		}`, string(data2))
	}

	m2 := NewCaseInsensitiveMapWithCapacity[string](m.Len())
	err = json.Unmarshal(data, m2)
	if assert.NoError(t, err) {
		val, ok = m2.Get("baz")
		assert.True(t, ok)
		assert.Equal(t, "qux", val)
	}
}
