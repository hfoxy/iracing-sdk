package irsdk

import "testing"

func TestInit(t *testing.T) {
	var sdk SDK
	sdk = Init(nil)
	sdk.Close()
}
