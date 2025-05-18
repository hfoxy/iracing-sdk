package buf

import "fmt"

var ErrWriterNotSet = fmt.Errorf("writer is not set")

func (s *Buffer) WriteVar(value int64) error {
	if s.writer == nil {
		return ErrWriterNotSet
	}

	var buffer []byte
	for {
		b := byte(value & 0x7F)
		value = value >> 7
		if value != 0 {
			b = b | 0x80
		}

		buffer = append(buffer, b)

		if value == 0 {
			break
		}
	}

	_, err := s.writer.Write(buffer)
	return err
}

func (s *Buffer) WriteVarInt(value int) error {
	return s.WriteVar(int64(value))
}

func (s *Buffer) WriteVarLong(value int64) error {
	return s.WriteVar(value)
}

func (s *Buffer) WriteBitMap(value BitMap) error {
	if s.writer == nil {
		return ErrWriterNotSet
	}

	v := byte(0)
	for i, b := range value {
		if b {
			v |= 1 << i
		}
	}

	n, err := s.writer.Write([]byte{v})
	if err != nil {
		return err
	} else if n != 1 {
		return fmt.Errorf("failed to write bitmap: %d", n)
	}

	return nil
}

func (s *Buffer) WriteString(value string) error {
	if s.writer == nil {
		return ErrWriterNotSet
	}

	err := s.WriteVarInt(len(value))
	if err != nil {
		return err
	}

	n, err := s.writer.Write([]byte(value))
	if err != nil {
		return err
	} else if n != len(value) {
		return fmt.Errorf("failed to write string: %d", n)
	}

	return nil
}

func (s *Buffer) Flush() error {
	if s.writer == nil {
		return ErrWriterNotSet
	}

	if f, ok := s.writer.(interface{ Flush() error }); ok {
		return f.Flush()
	} else {
		return fmt.Errorf("writer does not implement Flush")
	}

	return nil
}
