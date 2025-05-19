//go:build windows

package irsdk

import (
	"fmt"
	"time"

	"github.com/hfoxy/iracing-sdk/pkg/winevents"
)

// IRSDK is the main SDK object clients must use
type IRSDK struct {
	SDK
	r             reader
	h             *header
	s             string
	tVars         *TelemetryVars
	lastValidData int64
}

func (sdk *IRSDK) init() error {
	h, err := readHeader(sdk.r)
	if err != nil {
		return err
	}

	sdk.h = &h
	sdk.s = ""
	if sdk.tVars != nil {
		sdk.tVars.vars = nil
	}

	if sdk.sessionStatusOK() {
		err = sdk.RefreshSession()
		if err != nil {
			return err
		}

		var tVars *TelemetryVars
		tVars, err = readVariableHeaders(sdk.r, &h)
		if err != nil {
			return err
		}

		sdk.tVars = tVars

		_, err = sdk.readVariableValues()
		if err != nil {
			return err
		}
	}

	return nil
}

func (sdk *IRSDK) RefreshSession() error {
	if sdk.sessionStatusOK() {
		sRaw, err := readSessionData(sdk.r, sdk.h)
		if err != nil {
			return err
		}

		sdk.s = sRaw
	}

	return nil
}

func (sdk *IRSDK) sessionStatusOK() bool {
	return (sdk.h.status & stConnected) > 0
}

func (sdk *IRSDK) WaitForData(timeout time.Duration) (bool, error) {
	if !sdk.IsConnected() {
		return false, sdk.init()
	}

	if winevents.WaitForSingleObject(timeout) {
		err := sdk.RefreshSession()
		if err != nil {
			return false, err
		}

		return sdk.readVariableValues()
	}

	return false, nil
}

func (sdk *IRSDK) GetVars() ([]Variable, error) {
	if !sdk.sessionStatusOK() {
		return make([]Variable, 0), fmt.Errorf("session is not active")
	}

	results := make([]Variable, len(sdk.tVars.vars))

	sdk.tVars.mux.Lock()
	defer sdk.tVars.mux.Unlock()

	idx := 0
	for _, variable := range sdk.tVars.vars {
		results[idx] = variable
		idx++
	}

	return results, nil
}

func (sdk *IRSDK) GetVar(name string) (Variable, error) {
	if !sdk.sessionStatusOK() {
		return Variable{}, fmt.Errorf("session is not active")
	}

	sdk.tVars.mux.Lock()
	defer sdk.tVars.mux.Unlock()

	if v, ok := sdk.tVars.vars[name]; ok {
		return v, nil
	}

	return Variable{}, fmt.Errorf("telemetry variable %q not found", name)
}

var ErrNoValue = fmt.Errorf("no value")

func (sdk *IRSDK) GetVarValue(name string) (any, error) {
	var r Variable
	var err error

	if r, err = sdk.GetVar(name); err == nil {
		if len(r.Values) >= 1 {
			return r.Values[0], nil
		}

		return nil, ErrNoValue
	}

	return r, err
}

func (sdk *IRSDK) GetVarValues(name string) (interface{}, error) {
	var r Variable
	var err error

	if r, err = sdk.GetVar(name); err == nil {
		return r.Values, nil
	}

	return r, err
}

func (sdk *IRSDK) GetLastVersion() int {
	if !sdk.sessionStatusOK() {
		return -1
	}
	sdk.tVars.mux.Lock()
	defer sdk.tVars.mux.Unlock()
	last := sdk.tVars.lastVersion
	return last
}

func (sdk *IRSDK) IsConnected() bool {
	if sdk.h != nil {
		if sdk.sessionStatusOK() && (sdk.lastValidData+connTimeout > time.Now().Unix()) {
			return true
		}
	}

	return false
}

func (sdk *IRSDK) GetYaml() string {
	return sdk.s
}

func (sdk *IRSDK) BroadcastMsg(msg Msg) error {
	if msg.P2 == nil {
		msg.P2 = 0
	}

	_, err := winevents.BroadcastMsg(broadcastMsgName, msg.Cmd, msg.P1, msg.P2, msg.P3)
	return err
}

// Close clean up sdk resources
func (sdk *IRSDK) Close() error {
	return sdk.r.Close()
}
