package main

import (
	"context"
	irsdk "github.com/hfoxy/iracing-sdk/pkg"
	"log/slog"
	"time"
)

// TODO: this didn't copy over well from iTelemetry - assume it doesn't work for the time being

func main() {
	logger := slog.Default()
	sdk, err := irsdk.NewMock(irsdk.MockOptions{
		Logger:         logger.With("module", "irsdk"),
		DataSourceName: "session.zsitrpy",
	})

	if err != nil {
		panic(err)
	}

	defer sdk.Close()

	logger = logger.With("module", "example")
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

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
				v, err = sdk.GetVarValue("Speed")
				logger.Info("data", "value", v)
			}
		case <-ctx.Done():
			logger.Error("context done", "error", err)
			return
		}
	}
}
