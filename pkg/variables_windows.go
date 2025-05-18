package irsdk

import (
	"fmt"
	"sync"
	"time"
)

type VarBuffer struct {
	TickCount int // used to detect changes in data
	bufOffset int // offset from header
}

// TelemetryVars holds all variables we can read from telemetry live
type TelemetryVars struct {
	lastVersion int
	vars        map[string]Variable
	mux         sync.Mutex
}

func findLatestBuffer(r reader, h *header) (VarBuffer, error) {
	var vb VarBuffer
	foundTickCount := 0
	for i := 0; i < h.numBuf; i++ {
		rbuf := make([]byte, 16)
		_, err := r.ReadAt(rbuf, int64(48+i*16))
		if err != nil {
			return VarBuffer{}, err
		}

		currentVb := VarBuffer{
			byte4ToInt(rbuf[0:4]),
			byte4ToInt(rbuf[4:8]),
		}

		if foundTickCount < currentVb.TickCount {
			foundTickCount = currentVb.TickCount
			vb = currentVb
		}
	}

	return vb, nil
}

func readVariableHeaders(r reader, h *header) (*TelemetryVars, error) {
	vars := TelemetryVars{vars: make(map[string]Variable, h.numVars)}
	for i := 0; i < h.numVars; i++ {
		rbuf := make([]byte, 144)
		_, err := r.ReadAt(rbuf, int64(h.headerOffset+i*144))
		if err != nil {
			return nil, err
		}

		v := Variable{
			VarType:     VarType(byte4ToInt(rbuf[0:4])),
			Offset:      byte4ToInt(rbuf[4:8]),
			Count:       byte4ToInt(rbuf[8:12]),
			CountAsTime: int(rbuf[12]) > 0,
			Name:        bytesToString(rbuf[16:48]),
			Desc:        bytesToString(rbuf[48:112]),
			Unit:        bytesToString(rbuf[112:144]),
		}
		vars.vars[v.Name] = v
	}

	return &vars, nil
}

func (sdk *IRSDK) readVariableValues() (bool, error) {
	newData := false
	if sdk.sessionStatusOK() {
		// find latest buffer for variables
		vb, err := findLatestBuffer(sdk.r, sdk.h)
		if err != nil {
			return false, err
		}

		sdk.tVars.mux.Lock()
		if sdk.tVars.lastVersion < vb.TickCount {
			newData = true
			sdk.tVars.lastVersion = vb.TickCount
			sdk.lastValidData = time.Now().Unix()
			for varName, v := range sdk.tVars.vars {
				var rbuf []byte
				switch v.VarType {
				case VarTypeChar:
					values := make([]any, v.Count)
					for i := 0; i < v.Count; i++ {
						rbuf = make([]byte, 1)
						_, err = sdk.r.ReadAt(rbuf, int64(vb.bufOffset+v.Offset+(1*i)))
						if err != nil {
							return false, err
						}

						values[i] = string(rbuf[0])
					}

					v.Values = values
				case VarTypeBool:
					values := make([]any, v.Count)
					for i := 0; i < v.Count; i++ {
						rbuf = make([]byte, 1)
						_, err = sdk.r.ReadAt(rbuf, int64(vb.bufOffset+v.Offset+(1*i)))
						if err != nil {
							return false, err
						}

						values[i] = int(rbuf[0]) > 0
					}

					v.Values = values
				case VarTypeInt:
					values := make([]any, v.Count)
					for i := 0; i < v.Count; i++ {
						rbuf = make([]byte, 4)
						_, err = sdk.r.ReadAt(rbuf, int64(vb.bufOffset+v.Offset+(4*i)))
						if err != nil {
							return false, err
						}

						values[i] = byte4ToInt(rbuf)
					}

					v.Values = values
				case VarTypeBitField:
					values := make([]any, v.Count)
					for i := 0; i < v.Count; i++ {
						rbuf = make([]byte, 4)
						_, err = sdk.r.ReadAt(rbuf, int64(vb.bufOffset+v.Offset+(4*i)))
						if err != nil {
							return false, err
						}

						values[i] = byte4ToInt(rbuf)
					}

					v.Values = values
				case VarTypeFloat:
					values := make([]any, v.Count)
					for i := 0; i < v.Count; i++ {
						rbuf = make([]byte, 4)
						_, err = sdk.r.ReadAt(rbuf, int64(vb.bufOffset+v.Offset+(4*i)))
						if err != nil {
							return false, err
						}

						values[i] = byte4ToFloat(rbuf)
					}

					v.Values = values
				case VarTypeDouble:
					values := make([]any, v.Count)
					for i := 0; i < v.Count; i++ {
						rbuf = make([]byte, 8)
						_, err = sdk.r.ReadAt(rbuf, int64(vb.bufOffset+v.Offset+(8*i)))
						if err != nil {
							return false, err
						}

						values[i] = byte8ToFloat(rbuf)
					}

					v.Values = values
				default:
					return false, fmt.Errorf("unknown var type %d", v.VarType)
				}

				sdk.tVars.vars[varName] = v
			}
		}

		sdk.tVars.mux.Unlock()
	}

	return newData, nil
}
