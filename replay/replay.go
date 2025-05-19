package replay

import (
	"fmt"
	"os"
	"path/filepath"
)

var ErrEndOfFile = fmt.Errorf("end of file")

type Reader interface {
	ReadSize() int64
	Size() int64
	ReadEntry() (*Entry, error)
	Close() error
}

type Writer interface {
	WriteEntry(entry *Entry) error
	Close() error
}

type Entry struct {
	Timestamp    int64
	Connected    bool
	NotOk        bool
	YamlData     string
	VariableData string
}

func NewReader(fileName string) (Reader, error) {
	ext := filepath.Ext(fileName)

	// Get file size
	fileInfo, err := os.Stat(fileName)
	var fileSize int64
	if err == nil {
		fileSize = fileInfo.Size()
	}

	switch ext {
	case ".itrpy":
		return NewTelemetryReplayReader(fileName, fileSize)
	case ".zsitrpy":
		return NewTelemetryReplayReader(fileName, fileSize)
	case ".gzitrpy":
		return NewTelemetryReplayReader(fileName, fileSize)
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}
}

func NewWriter(fileName string) (Writer, error) {
	ext := filepath.Ext(fileName)
	switch ext {
	case ".itrpy":
		return NewTelemetryReplayWriter(fileName)
	case ".zsitrpy":
		return NewTelemetryReplayWriter(fileName)
	case ".gzitrpy":
		return NewTelemetryReplayWriter(fileName)
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}
}
