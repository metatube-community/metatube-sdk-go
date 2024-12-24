package bufferpool

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBufferPool(t *testing.T) {
	const size = 1024
	bp := New(size)
	for i := 0; i < 10; i++ {
		buf := bp.Get()
		assert.NotNil(t, buf)
		assert.Equal(t, buf.Len(), 0)
		assert.GreaterOrEqual(t, buf.Cap(), size)
		buf.WriteString(strings.Repeat("\x00", size*2))
		bp.Put(buf)
	}
}
