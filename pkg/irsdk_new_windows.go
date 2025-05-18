//go:build windows

package irsdk

import (
	"github.com/hfoxy/iracing-sdk/pkg/winevents"
	"github.com/hidez8891/shm"
)

// New creates SDK instance to operate with
func New() (SDK, error) {
	r, err := shm.Open(fileMapName, fileMapSize)
	if err != nil {
		return nil, err
	}

	sdk := &IRSDK{r: r, lastValidData: 0}
	winevents.OpenEvent(dataValidEventName)
	err = sdk.init()
	if err != nil {
		return nil, err
	}

	return sdk, nil
}
