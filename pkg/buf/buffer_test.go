package buf

import (
	"bytes"
	"compress/gzip"
	"errors"
	"github.com/klauspost/compress/zstd"
	"os"
	"testing"
)

func TestVarIntError(t *testing.T) {
	b := New(nil, nil, nil)
	err := b.WriteVarInt(0)
	if !errors.Is(err, ErrWriterNotSet) {
		t.Error("expected writer not set error")
	}

	_, _, err = b.ReadVarInt()
	if !errors.Is(err, ErrReaderNotSet) {
		t.Error("expected Reader not set error")
	}
}

func TestVarLongError(t *testing.T) {
	b := New(nil, nil, nil)
	err := b.WriteVarLong(0)
	if !errors.Is(err, ErrWriterNotSet) {
		t.Error("expected writer not set error")
	}

	_, _, err = b.ReadVarLong()
	if !errors.Is(err, ErrReaderNotSet) {
		t.Error("expected Reader not set error")
	}
}

func TestBitMapError(t *testing.T) {
	b := New(nil, nil, nil)
	err := b.WriteBitMap(BitMap{})
	if !errors.Is(err, ErrWriterNotSet) {
		t.Error("expected writer not set error")
	}

	_, err = b.ReadBitMap()
	if !errors.Is(err, ErrReaderNotSet) {
		t.Error("expected Reader not set error")
	}
}

func TestStringError(t *testing.T) {
	b := New(nil, nil, nil)
	err := b.WriteString("")
	if !errors.Is(err, ErrWriterNotSet) {
		t.Error("expected writer not set error")
	}

	_, err = b.ReadString()
	if !errors.Is(err, ErrReaderNotSet) {
		t.Error("expected Reader not set error")
	}
}

func TestVarInt(t *testing.T) {
	samples := []int32{0, 1, 127, 128, 255, 2147483647}

	buf := new(bytes.Buffer)
	b := New(buf, buf, nil)
	for _, sample := range samples {
		err := b.WriteVarInt(int(sample))
		if err != nil {
			t.Error(err)
			return
		}
	}

	for _, sample := range samples {
		v, _, err := b.ReadVarInt()
		if err != nil {
			t.Error(err)
			return
		}

		if v != sample {
			t.Errorf("expected %d, got %d", sample, v)
			return
		}
	}
}

func TestVarLong(t *testing.T) {
	samples := []int64{0, 1, 127, 128, 255, 2147483647, 2147483648, 9223372036854775807}

	buf := new(bytes.Buffer)
	b := New(buf, buf, nil)
	for _, sample := range samples {
		err := b.WriteVarLong(sample)
		if err != nil {
			t.Error(err)
			return
		}
	}

	for _, sample := range samples {
		v, _, err := b.ReadVarLong()
		if err != nil {
			t.Error(err)
			return
		}

		if v != sample {
			t.Errorf("expected %d, got %d", sample, v)
			return
		}
	}
}

func TestBitMap(t *testing.T) {
	samples := []BitMap{
		{true, false, true, false, true, false, true, false},
		{false, true, false, true, false, true, false, true},
		{false, false, true, true, false, false, true, true},
		{true, false, false, false, false, false, false, false},
	}

	buf := new(bytes.Buffer)
	b := New(buf, buf, nil)
	for _, sample := range samples {
		err := b.WriteBitMap(sample)
		if err != nil {
			t.Error(err)
			return
		}
	}

	for _, sample := range samples {
		v, err := b.ReadBitMap()
		if err != nil {
			t.Error(err)
			return
		}

		for i := 0; i < 8; i++ {
			if v[i] != sample[i] {
				t.Errorf("expected %v, got %v", sample, v)
				return
			}
		}
	}
}

func TestString(t *testing.T) {
	str := []string{"hello", "world", ""}

	buf := new(bytes.Buffer)
	b := New(buf, buf, nil)
	for _, s := range str {
		err := b.WriteString(s)
		if err != nil {
			t.Error(err)
			return
		}
	}

	for _, s := range str {
		v, err := b.ReadString()
		if err != nil {
			t.Error(err)
			return
		}

		if v != s {
			t.Errorf("expected %s, got %s", s, v)
			return
		}
	}
}

func TestStringSample(t *testing.T) {
	sampleData, err := os.ReadFile("sample.txt")
	if err != nil {
		t.Error(err)
		return
	}

	buf := new(bytes.Buffer)
	b := New(buf, buf, nil)
	err = b.WriteString(string(sampleData))
	if err != nil {
		t.Error(err)
		return
	}

	v, err := b.ReadString()
	if err != nil {
		t.Error(err)
		return
	}

	if string(sampleData) != v {
		t.Errorf("expected %s, got %s", string(sampleData), v)
	}
}

func TestStringSampleCompressedGzip(t *testing.T) {
	sampleData, err := os.ReadFile("sample.txt")
	if err != nil {
		t.Error(err)
		return
	}

	buf := new(bytes.Buffer)

	enc := gzip.NewWriter(buf)

	b := New(enc, nil, nil)
	err = b.WriteString(string(sampleData))
	if err != nil {
		t.Error(err)
		enc.Close()
		return
	}

	enc.Close()

	dec, err := gzip.NewReader(bytes.NewBuffer(buf.Bytes()))
	if err != nil {
		t.Error(err)
		return
	}

	defer dec.Close()

	b = New(nil, dec, nil)

	v, err := b.ReadString()
	if err != nil {
		t.Error(err)
		return
	}

	if string(sampleData) != v {
		t.Errorf("expected\n%s\n\ngot\n%s", string(sampleData), v)
		return
	}
}

func TestStringSampleCompressedZstd(t *testing.T) {
	sampleData, err := os.ReadFile("sample.txt")
	if err != nil {
		t.Error(err)
		return
	}

	buf := new(bytes.Buffer)

	enc, err := zstd.NewWriter(buf)
	if err != nil {
		t.Error(err)
		return
	}

	b := New(enc, nil, nil)
	err = b.WriteString(string(sampleData))
	if err != nil {
		t.Error(err)
		enc.Close()
		return
	}

	enc.Close()

	dec, err := zstd.NewReader(bytes.NewBuffer(buf.Bytes()))
	if err != nil {
		t.Error(err)
		return
	}

	defer dec.Close()

	b = New(nil, dec, nil)

	v, err := b.ReadString()
	if err != nil {
		t.Error(err)
		return
	}

	if string(sampleData) != v {
		t.Errorf("expected\n%s\n\ngot\n%s", string(sampleData), v)
		return
	}
}
