package irsdk

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"github.com/hfoxy/iracing-sdk/iryaml"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/hfoxy/iracing-sdk/lib/winevents"
	"github.com/hidez8891/shm"
)

type SDK interface {
	WaitForData(timeout time.Duration) bool
	GetVars() ([]Variable, error)
	GetVar(name string) (Variable, error)
	GetVarValue(name string) (interface{}, error)
	GetVarValues(name string) (interface{}, error)
	GetSession() iryaml.Session
	GetLastVersion() int
	IsConnected() bool
	ExportIbtTo(fileName string)
	ExportSessionTo(fileName string)
	BroadcastMsg(msg Msg)
	Close()
}

// IRSDK is the main SDK object clients must use
type IRSDK struct {
	SDK
	r             reader
	h             *header
	session       iryaml.Session
	s             []string
	tVars         *TelemetryVars
	lastValidData int64
}

func (sdk *IRSDK) WaitForData(timeout time.Duration) bool {
	if !sdk.IsConnected() {
		initIRSDK(sdk)
	}
	if winevents.WaitForSingleObject(timeout) {
		return readVariableValues(sdk)
	}
	return false
}

func (sdk *IRSDK) GetVars() ([]Variable, error) {
	if !sessionStatusOK(sdk.h.status) {
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
	if !sessionStatusOK(sdk.h.status) {
		return Variable{}, fmt.Errorf("session is not active")
	}

	sdk.tVars.mux.Lock()
	defer sdk.tVars.mux.Unlock()

	if v, ok := sdk.tVars.vars[name]; ok {
		return v, nil
	}

	return Variable{}, fmt.Errorf("Telemetry variable %q not found", name)
}

func (sdk *IRSDK) GetVarValue(name string) (interface{}, error) {
	var r Variable
	var err error

	if r, err = sdk.GetVar(name); err == nil {
		return r.Value, nil
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

func (sdk *IRSDK) GetSession() iryaml.Session {
	return sdk.session
}

func (sdk *IRSDK) GetLastVersion() int {
	if !sessionStatusOK(sdk.h.status) {
		return -1
	}
	sdk.tVars.mux.Lock()
	defer sdk.tVars.mux.Unlock()
	last := sdk.tVars.lastVersion
	return last
}

func (sdk *IRSDK) GetSessionData(path string) (string, error) {
	if !sessionStatusOK(sdk.h.status) {
		return "", fmt.Errorf("session not connected")
	}
	return getSessionDataPath(sdk.s, path)
}

func (sdk *IRSDK) IsConnected() bool {
	if sdk.h != nil {
		if sessionStatusOK(sdk.h.status) && (sdk.lastValidData+connTimeout > time.Now().Unix()) {
			return true
		}
	}

	return false
}

// ExportIbtTo exports current memory data to a file
func (sdk *IRSDK) ExportIbtTo(fileName string) {
	rbuf := make([]byte, fileMapSize)
	_, err := sdk.r.ReadAt(rbuf, 0)
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile(fileName, rbuf, 0644)
}

// ExportSessionTo exports current session yaml data to a file
func (sdk *IRSDK) ExportSessionTo(fileName string) {
	y := strings.Join(sdk.s, "\n")
	ioutil.WriteFile(fileName, []byte(y), 0644)
}

func (sdk *IRSDK) BroadcastMsg(msg Msg) {
	if msg.P2 == nil {
		msg.P2 = 0
	}
	winevents.BroadcastMsg(broadcastMsgName, msg.Cmd, msg.P1, msg.P2, msg.P3)
}

// Close clean up sdk resources
func (sdk *IRSDK) Close() {
	sdk.r.Close()
}

// Init creates a SDK instance to operate with
func Init(r reader) IRSDK {
	if r == nil {
		var err error
		r, err = shm.Open(fileMapName, fileMapSize)
		if err != nil {
			log.Fatal(err)
		}
	}

	sdk := IRSDK{r: r, lastValidData: 0}
	winevents.OpenEvent(dataValidEventName)
	initIRSDK(&sdk)
	return sdk
}

func initIRSDK(sdk *IRSDK) {
	h := readHeader(sdk.r)
	sdk.h = &h
	sdk.s = nil
	if sdk.tVars != nil {
		sdk.tVars.vars = nil
	}
	if sessionStatusOK(h.status) {
		sRaw := readSessionData(sdk.r, &h)
		err := yaml.Unmarshal([]byte(sRaw), &sdk.session)
		if err != nil {
			log.Fatal(err)
		}
		sdk.s = strings.Split(sRaw, "\n")
		sdk.tVars = readVariableHeaders(sdk.r, &h)
		readVariableValues(sdk)
	}
}

func sessionStatusOK(status int) bool {
	return (status & stConnected) > 0
}
