package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
)

type Report struct {
	UTC          int64   `json:"utc"`
	EngineSpeed  float64 `json:"engine_speed"`
	VehicleSpeed int     `json:"vehicle_speed"`
}

func main() {
	// Initialize BLE device
	d, err := dev.NewDevice("default")
	if err != nil {
		log.Fatalf("Failed to initialize BLE device: %v", err)
	}
	ble.SetDefaultDevice(d)

	elm327ble, err := NewELM327BLE(true)
	if err != nil {
		log.Fatalf("Failed to create ELM327BLE: %v", err)
	}
	defer elm327ble.Close()

	// Receive Ctrl+C signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	tk := time.NewTicker(500 * time.Millisecond)
	defer tk.Stop()
	for {
		select {
		case <-sigs:
			return
		case <-tk.C:
			vs, err := elm327ble.ReadCurrDatas(
				[]byte{OBD2Svc01EngineSpeed, OBD2Svc01VehicleSpeed},
			)
			if err != nil {
				log.Printf("Failed to read current datas: %v", err)
				continue
			}

			var r Report
			r.UTC = time.Now().UnixMilli()
			for _, v := range vs {
				switch v.Pid {
				case OBD2Svc01EngineSpeed:
					r.EngineSpeed = float64(256*int(v.Data[0])+int(v.Data[1])) / 4
				case OBD2Svc01VehicleSpeed:
					r.VehicleSpeed = int(v.Data[0])
				}
			}
			jsonStr, _ := json.Marshal(r)
			log.Printf("%s", jsonStr)
		}
	}
}
