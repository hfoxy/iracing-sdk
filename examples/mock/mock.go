package main

import (
	"context"
	"fmt"
	irsdk "github.com/hfoxy/iracing-sdk/pkg"
	"log/slog"
	"time"
)

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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

			fmt.Printf("data received: %t/%t\n", ok, sdk.IsConnected())
		case <-ctx.Done():
			logger.Error("context done", "error", err)
			return
		}
	}
}
