package buf

import (
	"errors"
	"fmt"
	"io"
)

var ErrReaderNotSet = fmt.Errorf("reader is not set")

func (s *Buffer) ReadCount() int64 {
	if s.readCount == nil {
		return -1
	} else {
		return *s.readCount
	}
}

func (s *Buffer) ReadVar(maxWidth int, lengthStringErr string) (int64, int, error) {
	if s.reader == nil {
		return -1, 0, ErrReaderNotSet
	}

	numRead := 0
	result := int64(0)
	read := make([]byte, 1)

	for {
		_, err := s.reader.Read(read)
		if err != nil {
			return -1, numRead, err
		}

		result |= int64(read[0]&0x7F) << uint(7*numRead)
		numRead++

		if (read[0] & 0x80) == 0 {
			break
		}

		if numRead >= maxWidth {
			return -1, numRead, fmt.Errorf("%s: %d >= %d (%b)", lengthStringErr, numRead, maxWidth, result)
		}
	}

	return result, numRead, nil
}

func (s *Buffer) ReadVarInt() (int32, int, error) {
	v, n, err := s.ReadVar(5, "varint too long")
	return int32(v), n, err
}

func (s *Buffer) ReadVarLong() (int64, int, error) {
	return s.ReadVar(10, "varlong too long")
}

func (s *Buffer) ReadBitMap() (BitMap, error) {
	if s.reader == nil {
		return BitMap{}, ErrReaderNotSet
	}

	data := make([]byte, 1)
	_, err := s.reader.Read(data)
	if err != nil {
		return BitMap{}, err
	}

	v := data[0]
	b := BitMap{}
	for i := 0; i < 8; i++ {
		b[i] = v&(1<<i) != 0
	}

	return b, nil
}

var chunkSize = 4096

func (s *Buffer) ReadString() (string, error) {
	if s.reader == nil {
		return "", ErrReaderNotSet
	}

	length, _, err := s.ReadVarInt()
	if err != nil {
		return "", err
	}

	read := 0
	data := make([]byte, 0, length)

	for read < int(length) {
		ch := chunkSize
		if read+chunkSize > int(length) {
			ch = int(length) - read
		}

		if ch == 0 {
			break
		}

		var n int
		sl := make([]byte, ch)
		n, err = s.reader.Read(sl)
		if err != nil {
			if n != ch || !errors.Is(err, io.EOF) {
				return string(data), fmt.Errorf("unable to read string (%d / %d / %d): %w", len(sl), read, n, err)
			}
		}

		data = append(data, sl[:n]...)

		read += n

		if n == 0 || read == int(length) {
			// log.Printf("break ch(%d) (%d == 0 || %d == %d)\n", ch, n, read, length)
			break
		}
	}

	return string(data), nil
}
