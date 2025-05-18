package replay

import (
	"compress/gzip"
	"fmt"
	"github.com/hfoxy/iracing-sdk/pkg/buf"
	"github.com/klauspost/compress/zstd"
	"io"
	"os"
	"path/filepath"
)

type TelemetryReplayWriter struct {
	closeFunc func() error
	buf       *buf.Buffer

	lastYamlData     string
	lastVariableData string
}

func NewTelemetryReplayWriter(outputFile string) (Writer, error) {
	if _, err := os.Stat(outputFile); err == nil {
		return nil, fmt.Errorf("output file already exists: %s", outputFile)
	}

	f, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open output file: %s", outputFile)
	}

	var closeFunc func() error
	var w io.Writer

	ext := filepath.Ext(outputFile)
	switch ext {
	case ".itrpy":
		closeFunc = f.Close
		w = f
	case ".zsitrpy":
		var enc *zstd.Encoder
		enc, err = zstd.NewWriter(f)
		if err != nil {
			return nil, fmt.Errorf("failed to create zstd encoder: %w", err)
		}

		closeFunc = func() error {
			enc.Close()
			return f.Close()
		}

		w = enc
	case ".gzitrpy":
		enc := gzip.NewWriter(f)
		closeFunc = enc.Close
		w = enc
	default:
		return nil, fmt.Errorf("unsupported format: %s", ext)
	}

	return &TelemetryReplayWriter{
		closeFunc: closeFunc,
		buf:       buf.NewWriter(w),
	}, nil
}

func (w *TelemetryReplayWriter) WriteEntry(entry *Entry) error {
	yamlUpdated := w.lastYamlData != entry.YamlData
	variableUpdated := w.lastVariableData != entry.VariableData

	bm := buf.BitMap{yamlUpdated, variableUpdated, entry.Connected, entry.NotOk}

	err := w.buf.WriteVarLong(entry.Timestamp)
	if err != nil {
		return err
	}

	err = w.buf.WriteBitMap(bm)
	if err != nil {
		return err
	}

	if yamlUpdated {
		err = w.buf.WriteString(entry.YamlData)
		if err != nil {
			return err
		}
	}

	if variableUpdated {
		err = w.buf.WriteString(entry.VariableData)
		if err != nil {
			return err
		}
	}

	w.lastYamlData = entry.YamlData
	w.lastVariableData = entry.VariableData
	return w.buf.Flush()
}

func (w *TelemetryReplayWriter) Close() error {
	if w.closeFunc == nil {
		return nil
	}

	return w.closeFunc()
}
