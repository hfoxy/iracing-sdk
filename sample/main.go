package main

import (
	"errors"
	irsdk "github.com/hfoxy/iracing-sdk"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var sdk irsdk.SDK

func main() {
	gracefulShutdown := make(chan os.Signal, 1)
	signal.Notify(gracefulShutdown, syscall.SIGINT, syscall.SIGTERM)

	log.Println("sample starting...")

	sdk = irsdk.Init(nil)

	ticker := time.NewTicker(100 * time.Millisecond)
	quit := make(chan struct{})
	for {
		select {
		case <-quit:
		case <-gracefulShutdown:
			ticker.Stop()
			log.Println("shutdown requested by CTRL+C")
			sdk.Close()
			return
		case <-ticker.C:
			sdk.WaitForData(150 * time.Millisecond)
			if !sdk.IsConnected() {
				continue
			}

			if _, err := os.Stat("session.yaml"); errors.Is(err, os.ErrNotExist) {
				sdk.ExportSessionTo("session.yaml")
			}

			va, _ := sdk.GetVar("Speed")
			log.Printf("speed: %#v (%s)", va.Value, va.Unit)

			v, _ := sdk.GetVarValue("SessionFlags")
			log.Printf("flags: %#v", v)

			v, _ = sdk.GetVarValue("PitSvFlags")
			log.Printf("pit sv flags: %#v", v)

			v, _ = sdk.GetVarValues("CarIdxClassPosition")
			log.Printf("class position: %#v", v)

			//for _, driver := range sdk.GetSession().DriverInfo.Drivers {
			//	logging.Logger.Infof("Driver: %s", driver.UserName)
			//}
		}
	}
}
