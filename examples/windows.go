package main

import (
	"context"
	"github.com/hfoxy/iracing-sdk"
	"log/slog"
	"time"
)

func main() {
	logger := slog.Default()
	sdk, err := irsdk.New()
	if err != nil {
		panic(err)
	}

	defer sdk.Close()

	logger = logger.With("module", "example")
	ctx := context.Background()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var ok bool
	for {
		select {
		case <-ticker.C:
			ok, err = sdk.WaitForData(50 * time.Millisecond)
			if err != nil {
				logger.Error("unable to wait for data", "error", err)
				return
			}

			logger.Info("data received", "ok", ok, "connected", sdk.IsConnected())

			if ok && sdk.IsConnected() {
				var v interface{}
				v, err = sdk.GetVarValue("SessionTime")
				logger.Info("data", "value", v)
			}
		case <-ctx.Done():
			logger.Error("context done", "error", err)
			return
		}
	}
}
