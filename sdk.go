package irsdk

import (
	"fmt"
	"time"
)

var ErrNotImplemented = fmt.Errorf("not implemented - placeholder")

type SDK interface {
	WaitForData(timeout time.Duration) (bool, error)
	GetVars() ([]Variable, error)
	GetVar(name string) (Variable, error)
	GetVarValue(name string) (interface{}, error)
	GetVarValues(name string) (interface{}, error)
	RefreshSession() error
	GetLastVersion() int
	IsConnected() bool
	GetYaml() string
	BroadcastMsg(msg Msg) error
	Close() error
}
