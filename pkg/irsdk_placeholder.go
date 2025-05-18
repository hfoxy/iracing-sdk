//go:build !mock && !windows

package irsdk

import (
	"fmt"
	"time"
)

type IRSDK struct {
}

func New() (SDK, error) {
	return &IRSDK{}, nil
}

var ErrNotImplemented = fmt.Errorf("not implemented - placeholder")

func (sdk *IRSDK) RefreshSession() error {
	//TODO implement me
	return ErrNotImplemented
}

func (sdk *IRSDK) WaitForData(timeout time.Duration) (bool, error) {
	return false, ErrNotImplemented
}

func (sdk *IRSDK) GetVars() ([]Variable, error) {
	return nil, ErrNotImplemented
}

func (sdk *IRSDK) GetVar(name string) (Variable, error) {
	return Variable{}, ErrNotImplemented
}

func (sdk *IRSDK) GetVarValue(name string) (interface{}, error) {
	return nil, ErrNotImplemented
}

func (sdk *IRSDK) GetVarValues(name string) (interface{}, error) {
	return nil, ErrNotImplemented
}

func (sdk *IRSDK) GetLastVersion() int {
	//TODO implement me
	return -1
}

func (sdk *IRSDK) IsConnected() bool {
	return false
}

func (sdk *IRSDK) GetYaml() string {
	return ""
}

func (sdk *IRSDK) BroadcastMsg(msg Msg) error {
	return ErrNotImplemented
}

func (sdk *IRSDK) Close() error {
	return ErrNotImplemented
}
