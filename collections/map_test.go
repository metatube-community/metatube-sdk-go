package collections

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderedMap(t *testing.T) {
	{ // JSON marshal
		m := NewOrderedMap[string, any]()
		b, _ := json.Marshal(m)
		require.Equal(t, `{}`, string(b))

		m.Set("a", 1)
		m.Set("c", "2")
		m.Set("b", 3.0)
		m.Set("b", 1.5)
		require.Equal(t, []any{1, "2", 1.5}, slices.Collect(m.Values()))

		b, _ = json.Marshal(m)
		require.Equal(t, `{"a":1,"c":"2","b":1.5}`, string(b))
	}

	{ // fixed type map unmarshal
		jsonData := `{
			"a":1,
			"c":2,
			"b":0
		}`
		m := NewOrderedMap[string, int]()
		err := m.UnmarshalJSON([]byte(jsonData))
		require.NoError(t, err)

		b, _ := json.Marshal(m)
		require.Equal(t, `{"a":1,"c":2,"b":0}`, string(b))
	}

	{ // any type map unmarshal
		jsonData := `{
			"a":1,
			"c":"2",
			"b":1.5,
			"?":{"x":"y","j":"k","3":2}
		}`
		m := NewOrderedMap[string, any]()
		err := m.UnmarshalJSON([]byte(jsonData))
		require.NoError(t, err)

		b, _ := json.Marshal(m)
		require.Equal(t, `{"a":1,"c":"2","b":1.5,"?":{"3":2,"j":"k","x":"y"}}`, string(b))
	}

	{ // Sorted sub map unmarshal
		jsonData := `{
			"w":{"n":3,"m":5},
			"b":{"f":1,"j":0}
		}`
		m := NewOrderedMap[string, map[string]int]()
		err := m.UnmarshalJSON([]byte(jsonData))
		require.NoError(t, err)

		b, _ := json.Marshal(m)
		require.Equal(t, `{"w":{"m":5,"n":3},"b":{"f":1,"j":0}}`, string(b))
	}

	{ // Ordered sub map unmarshal
		jsonData := `{
			"w":{"n":3,"m":5},
			"b":{"f":1,"j":0}
		}`
		m := NewOrderedMap[string, *OrderedMap[string, int]]()
		err := m.UnmarshalJSON([]byte(jsonData))
		require.NoError(t, err)

		b, _ := json.Marshal(m)
		require.Equal(t, `{"w":{"n":3,"m":5},"b":{"f":1,"j":0}}`, string(b))
	}

	{ // A lot of ordered sub maps unmarshal
		jsonData := `{
			"w":{"n":{"g":3,"5":5},"m":{"v":3,"2":5}},
			"b":{"f":{"h":3,"3":5},"j":{"x":3,"c":5}}
		}`
		m := NewOrderedMap[string, *OrderedMap[string, *OrderedMap[string, any]]]()
		err := m.UnmarshalJSON([]byte(jsonData))
		require.NoError(t, err)

		b, _ := json.Marshal(m)
		require.Equal(t, `{"w":{"n":{"g":3,"5":5},"m":{"v":3,"2":5}},"b":{"f":{"h":3,"3":5},"j":{"x":3,"c":5}}}`, string(b))
	}
}
