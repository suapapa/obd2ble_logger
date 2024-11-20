package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
)

func main() {
	var debug bool
	var dumpFile string

	flag.BoolVar(&debug, "d", false, "enable debug print")
	flag.StringVar(&dumpFile, "f", "", "write log to file")

	// Initialize BLE device
	d, err := dev.NewDevice("default")
	if err != nil {
		log.Fatalf("Failed to initialize BLE device: %v", err)
	}
	ble.SetDefaultDevice(d)

	elm327ble, err := NewELM327BLE(debug)
	if err != nil {
		log.Fatalf("Failed to create ELM327BLE: %v", err)
	}
	defer elm327ble.Close()

	// Receive Ctrl+C signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	tk := time.NewTicker(500 * time.Millisecond)
	defer tk.Stop()

	var f *os.File
	if dumpFile != "" {
		f, err = os.Create(dumpFile)
		if err != nil {
			log.Fatalf("fail to create dumpfile: %v", err)
		}
		defer f.Close()
	}

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

			// var r Report
			r := make(map[string]interface{})
			r["utc"] = time.Now().UnixMilli()
			for _, v := range vs {
				r[v.PidString()] = v.String()
			}
			jsonBytes, _ := json.Marshal(r)
			if f != nil {
				f.Write(jsonBytes)
				f.Write([]byte{'\n'})
				f.Sync()
			}
			fmt.Println(string(jsonBytes))
		}
	}
}
