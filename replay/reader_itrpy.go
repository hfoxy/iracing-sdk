package replay

import (
	"compress/gzip"
	"errors"
	"fmt"
	buf2 "github.com/hfoxy/iracing-sdk/buf"
	"github.com/klauspost/compress/zstd"
	"io"
	"os"
	"path/filepath"
)

type TelemetryReplayReader struct {
	fileName  string
	fileSize  int64
	closeFunc func() error
	reader    *buf2.Buffer

	yamlData     string
	variableData string
}

func NewTelemetryReplayReader(fileName string, fileSize int64) (Reader, error) {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", fileName)
	}

	f, err := os.OpenFile(fileName, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %s", fileName)
	}

	var closeFunc func() error

	var r io.Reader
	r = f

	cr := &buf2.CountingReader{
		Reader: r,
	}

	r = cr

	ext := filepath.Ext(fileName)
	switch ext {
	case ".itrpy":
		closeFunc = f.Close
	case ".zsitrpy":
		var dec *zstd.Decoder
		dec, err = zstd.NewReader(r)
		if err != nil {
			return nil, fmt.Errorf("failed to create zstd decoder: %w", err)
		}

		r = dec
		closeFunc = func() error {
			dec.Close()
			return f.Close()
		}
	case ".gzitrpy":
		var dec *gzip.Reader
		dec, err = gzip.NewReader(r)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip decoder: %w", err)
		}

		r = dec
		closeFunc = func() error {
			if err = dec.Close(); err != nil {
				return err
			}

			return f.Close()
		}
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	buffer := buf2.NewReader(r, &cr.BytesRead)

	return &TelemetryReplayReader{
		fileName:  fileName,
		fileSize:  fileSize,
		closeFunc: closeFunc,
		reader:    buffer,
	}, nil
}

func (s *TelemetryReplayReader) ReadSize() int64 {
	return s.reader.ReadCount()
}

func (s *TelemetryReplayReader) Size() int64 {
	return s.fileSize
}

func (s *TelemetryReplayReader) ReadEntry() (*Entry, error) {
	entry := &Entry{}

	var err error
	entry.Timestamp, _, err = s.reader.ReadVarLong()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return entry, ErrEndOfFile
		}

		return entry, fmt.Errorf("failed to read timestamp: %w", err)
	}

	bm, err := s.reader.ReadBitMap()
	if err != nil {
		return entry, fmt.Errorf("failed to read bitmap: %w", err)
	}

	yamlUpdated := bm[0]
	variableUpdated := bm[1]
	entry.Connected = bm[2]
	entry.NotOk = bm[3]

	if yamlUpdated {
		entry.YamlData, err = s.reader.ReadString()
		if err != nil {
			return entry, fmt.Errorf("failed to read yaml data: %w", err)
		}

		s.yamlData = entry.YamlData
	} else {
		entry.YamlData = s.yamlData
	}

	if variableUpdated {
		entry.VariableData, err = s.reader.ReadString()
		if err != nil {
			return entry, fmt.Errorf("failed to read variable data: %w", err)
		}

		s.variableData = entry.VariableData
	} else {
		entry.VariableData = s.variableData
	}

	return entry, nil
}

func (s *TelemetryReplayReader) Close() error {
	if s.closeFunc != nil {
		return s.closeFunc()
	}

	return nil
}
