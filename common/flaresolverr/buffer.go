package flaresolverr

import (
	"bytes"
	"io"
)

var _ io.ReadCloser = (*readCloser)(nil)

type readCloser struct {
	*bytes.Reader
}

func newReadCloser(b []byte) *readCloser {
	return &readCloser{
		Reader: bytes.NewReader(b),
	}
}

func newReadCloserString(s string) *readCloser {
	return newReadCloser([]byte(s))
}

func (b *readCloser) Close() error {
	return nil
}
