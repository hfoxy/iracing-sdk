package buf

import (
	"io"
)

type Buffer struct {
	writer    io.Writer
	reader    io.Reader
	readCount *int64
}

func New(writer io.Writer, reader io.Reader, readCount *int64) *Buffer {
	return &Buffer{
		writer:    writer,
		reader:    reader,
		readCount: readCount,
	}
}

func NewReader(reader io.Reader, readCount *int64) *Buffer {
	return New(nil, reader, readCount)
}

func NewWriter(writer io.Writer) *Buffer {
	return New(writer, nil, nil)
}

type BitMap [8]bool
