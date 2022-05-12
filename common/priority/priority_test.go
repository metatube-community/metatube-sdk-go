package priority

import (
	"bytes"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlice_Underlying(t *testing.T) {
	s := &Slice[int, byte]{}
	s.Append(2, 5)
	s.Append(3, 1)
	s.Append(1, 9)
	s.Append(6, 9)
	s.Append(0, 12)
	sort.Sort(s)
	assert.True(t, bytes.Equal(s.Underlying(), []byte{9, 1, 5, 9, 12}))
}
