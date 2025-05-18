package irsdk

import "io"

const dataValidEventName string = "Local\\IRSDKDataValidEvent"
const fileMapName string = "Local\\IRSDKMemMapFileName"
const fileMapSize int32 = 1164 * 1024
const broadcastMsgName string = "IRSDK_BROADCASTMSG"
const connTimeout = 30

const (
	stConnected int = 1
)

type reader interface {
	io.Reader
	io.ReaderAt
	io.Closer
}
