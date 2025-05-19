package irsdk

import (
	"strings"

	"golang.org/x/text/encoding/charmap"
)

func readSessionData(r reader, h *header) (string, error) {
	// session data (yaml)
	dec := charmap.Windows1252.NewDecoder()
	rbuf := make([]byte, h.sessionInfoLen)
	_, err := r.ReadAt(rbuf, int64(h.sessionInfoOffset))
	if err != nil {
		return "", err
	}

	rbuf, err = dec.Bytes(rbuf)
	if err != nil {
		return "", err
	}

	yaml := strings.TrimRight(string(rbuf[:h.sessionInfoLen]), "\x00")
	return yaml, nil
}
