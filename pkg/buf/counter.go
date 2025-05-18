package buf

import (
	"io"
)

type CountingReader struct {
	Reader    io.Reader
	BytesRead int64
}

func (r *CountingReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.BytesRead += int64(n)
	return n, err
}
