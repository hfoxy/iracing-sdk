//go:build windows

package irsdk

import "testing"

func TestInit(t *testing.T) {
	sdk, err := New()
	if err != nil {
		t.Fatal(err)
	}

	sdk.Close()
}
