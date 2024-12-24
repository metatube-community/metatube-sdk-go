package bufferpool

import (
	"bytes"

	"github.com/metatube-community/metatube-sdk-go/common/pool"
)

type BufferPool struct {
	pool *pool.Pool[*bytes.Buffer]
}

func New(size int) *BufferPool {
	return &BufferPool{
		pool: pool.New(func() *bytes.Buffer {
			return bytes.NewBuffer(make([]byte, 0, size))
		}),
	}
}

func (bp *BufferPool) Get() *bytes.Buffer {
	buf := bp.pool.Get()
	buf.Reset()
	return buf
}

func (bp *BufferPool) Put(b *bytes.Buffer) {
	bp.pool.Put(b)
}
