package priority

import (
	"bytes"
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
	assert.True(t, bytes.Equal(s.Sort().Underlying(), []byte{9, 1, 5, 9, 12}))
	assert.True(t, bytes.Equal(s.Reverse().Underlying(), []byte{12, 9, 5, 1, 9}))
}
