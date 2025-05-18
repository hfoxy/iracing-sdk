package irsdk

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/hfoxy/iracing-sdk/pkg/replay"
	"log/slog"
	"os"
	"time"
)

type MockSDK struct {
	logger  Logger
	options MockOptions

	replay replay.Reader

	openTime  time.Time
	startTime time.Time

	ended bool

	lastEntry     *replay.Entry
	lastEntryTime time.Time

	nextEntry     *replay.Entry
	nextEntryTime time.Time

	currentRow *row

	restartAllowedFrom time.Time
}

type MockOptions struct {
	Logger Logger

	DataSourceName   string
	AutoRestart      bool
	AutoRestartDelay time.Duration
	AutoRestartQuit  bool
}

func NewMock(opts MockOptions) (*MockSDK, error) {
	if opts.DataSourceName == "" {
		return nil, errors.New("data source name cannot be empty")
	}

	sdk := &MockSDK{}

	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}

	sdk.logger = opts.Logger

	var err error
	sdk.replay, err = replay.NewReader(opts.DataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open replay: %w", err)
	}

	sdk.lastEntry, err = sdk.replay.ReadEntry()
	if err != nil {
		return nil, fmt.Errorf("failed to read first entry: %w", err)
	}

	sdk.lastEntryTime = time.Unix(0, sdk.lastEntry.Timestamp*int64(time.Millisecond))

	sdk.nextEntry, err = sdk.replay.ReadEntry()
	if err != nil {
		return nil, fmt.Errorf("failed to read second entry: %w", err)
	}

	sdk.nextEntryTime = time.Unix(0, sdk.nextEntry.Timestamp*int64(time.Millisecond))

	sdk.startTime = sdk.lastEntryTime
	sdk.openTime = time.Now().Add(-5 * time.Second)
	return sdk, nil
}

func (sdk *MockSDK) RefreshSession() error {
	//TODO implement me
	return ErrNotImplemented
}

type row struct {
	Timestamp    int64
	Connected    bool
	NotOk        bool
	YamlData     string
	VariableData string
	Variables    []Variable
}

func (sdk *MockSDK) WaitForData(timeout time.Duration) (bool, error) {
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	t := sdk.startTime.Add(time.Now().Sub(sdk.openTime))
	millis := t.UnixMilli()

	if millis < sdk.nextEntry.Timestamp {
		return true, nil
	}

	r := row{}

	if !sdk.ended {
		end := false

		read := 0
		for {
			sdk.lastEntry = sdk.nextEntry
			sdk.lastEntryTime = sdk.nextEntryTime

			var err error
			sdk.nextEntry, err = sdk.replay.ReadEntry()
			if err != nil {
				if errors.Is(err, replay.ErrEndOfFile) {
					end = true
				} else {
					return false, fmt.Errorf("failed to get data: %w", err)
				}
			} else if sdk.nextEntry == nil {
				sdk.logger.Warn("next entry is nil")
				end = true
			}

			sdk.nextEntryTime = time.Unix(0, sdk.nextEntry.Timestamp*int64(time.Millisecond))
			read++

			if sdk.lastEntryTime.Before(t) && sdk.nextEntryTime.After(t) {
				break
			}

			r.Timestamp = sdk.lastEntry.Timestamp
			r.Connected = sdk.lastEntry.Connected
			r.NotOk = sdk.lastEntry.NotOk
			r.YamlData = sdk.lastEntry.YamlData
			r.VariableData = sdk.lastEntry.VariableData

			if end {
				break
			}
		}

		if end {
			sdk.ended = true
		}
	}

	sdk.currentRow = &r
	if sdk.ended && sdk.restartAllowedFrom.IsZero() {
		sdk.restartAllowedFrom = time.Now().Add(sdk.options.AutoRestartDelay)
		sdk.logger.Info("reached end of recording", "restartAllowedFrom", sdk.restartAllowedFrom.Format(time.RFC3339))
	}

	if sdk.options.AutoRestart && !sdk.restartAllowedFrom.IsZero() && time.Now().After(sdk.restartAllowedFrom) {
		if sdk.options.AutoRestartQuit {
			sdk.logger.Info("auto-restart quit enabled - quitting")
			// shutdown.Shutdown()
			os.Exit(0)
			return false, nil
		}

		return false, nil
	}

	if r.Connected {
		if r.VariableData != "" {
			vd, err := base64.StdEncoding.DecodeString(r.VariableData)
			if err != nil {
				return false, fmt.Errorf("failed to decode variable data: %w", err)
			}

			err = gob.NewDecoder(bytes.NewBuffer(vd)).Decode(&r.Variables)
			if err != nil {
				return false, fmt.Errorf("failed to decode variables: %w", err)
			}

			r.Variables = sdk.currentRow.Variables
		}

		return !r.NotOk, nil
	}

	return false, nil
}

func (sdk *MockSDK) GetVars() ([]Variable, error) {
	if sdk.currentRow == nil {
		return make([]Variable, 0), nil
	}

	return sdk.currentRow.Variables, nil
}

func (sdk *MockSDK) GetVar(name string) (Variable, error) {
	for _, variable := range sdk.currentRow.Variables {
		if variable.Name == name {
			return variable, nil
		}
	}

	return Variable{}, fmt.Errorf("variable not found: %s", name)
}

func (sdk *MockSDK) GetVarValue(name string) (interface{}, error) {
	v, err := sdk.GetVar(name)
	if err != nil {
		return nil, err
	}

	if len(v.Values) > 0 {
		return v.Values[0], nil
	}

	return nil, fmt.Errorf("variable entry not found: %s", name)
}

func (sdk *MockSDK) GetVarValues(name string) (interface{}, error) {
	v, err := sdk.GetVar(name)
	if err != nil {
		return nil, err
	}

	if v.Values != nil {
		return v.Values, nil
	} else {
		return nil, fmt.Errorf("variable entries not found: %s", name)
	}
}

func (sdk *MockSDK) GetLastVersion() int {
	//TODO implement me
	return -1
}

func (sdk *MockSDK) IsConnected() bool {
	if sdk.options.AutoRestart && !sdk.restartAllowedFrom.IsZero() && time.Now().After(sdk.restartAllowedFrom) {
		return false
	}

	if sdk.currentRow == nil {
		return false
	}

	return sdk.currentRow.Connected
}

func (sdk *MockSDK) GetYaml() string {
	if sdk.currentRow == nil {
		return ""
	}

	return sdk.currentRow.YamlData
}

func (sdk *MockSDK) BroadcastMsg(msg Msg) error {
	//TODO implement me
	return ErrNotImplemented
}

func (sdk *MockSDK) Close() error {
	if sdk.replay == nil {
		return nil
	}

	if err := sdk.replay.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	return nil
}
