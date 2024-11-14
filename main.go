package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
)

var (
	// Target device name or address
	targetDeviceName = "IOS-Vlink"

	// UUIDs for service and characteristic
	serviceUUID        = ble.MustParse("E7810A71-73AE-499D-8C15-FAA9AEF0C3F2")
	characteristicUUID = ble.MustParse("BEF8D6C9-9C21-4C9E-B632-BD58C1009F9F")

	client               ble.Client
	targetCharacteristic *ble.Characteristic
)

func main() {
	// Initialize BLE device
	d, err := dev.NewDevice("default")
	if err != nil {
		log.Fatalf("Failed to initialize BLE device: %v", err)
	}
	ble.SetDefaultDevice(d)

	// Scan for the target device
	log.Println("Scanning for BLE devices...")
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), 10*time.Second))
	client, err = ble.Connect(ctx, func(a ble.Advertisement) bool {
		if a.LocalName() == targetDeviceName {
			log.Printf("Found target device: %s(%s)", targetDeviceName, a.Addr())
			return true
		}
		return false
	})
	if err != nil {
		log.Fatalf("Failed to connect to device: %v", err)
	}
	defer client.CancelConnection()

	log.Printf("Connected to %s", targetDeviceName)

	// Discover the specified service
	log.Println("Discovering services...")
	profile, err := client.DiscoverProfile(true)
	if err != nil {
		log.Fatalf("Failed to discover profile: %v", err)
	}

	// var targetCharacteristic *ble.Characteristic
	for _, service := range profile.Services {
		if service.UUID.Equal(serviceUUID) {
			log.Println("Found target service")
			for _, char := range service.Characteristics {
				if char.UUID.Equal(characteristicUUID) {
					log.Println("Found target characteristic")
					targetCharacteristic = char
					break
				}
			}
		}
	}

	if targetCharacteristic == nil {
		log.Fatalf("Target characteristic not found")
	}

	respStrm := NewRespStream()
	err = client.Subscribe(targetCharacteristic, false, func(data []byte) {
		// log.Printf("notify: %x, %s", data, string(data))
		respStrm.Write(data)
	})
	if err != nil {
		log.Fatal("notify err")
	}

	// Commands to send
	commands := []string{
		"ATZ",
		"AT E0",
		"AT L0",
		"AT SP 00",
		"01 00",
	}
	// Write each command and read response
	for _, command := range commands {
		writeCmd(command, respStrm)
	}

	log.Println("Initialization complete")

	log.Println("finding supported commands")
	for i := 32; i < 0x81; i += 32 {
		command := fmt.Sprintf("01 %02x", i)
		writeCmd(command, respStrm)
	}
}

func writeCmd(command string, respStrm *RespStream) *Value {
	log.Printf("Writing command: %s", command)
	// cmd := []byte(command)
	cmd := append([]byte(command), '\r')
	if err := client.WriteCharacteristic(targetCharacteristic, cmd, false); err != nil {
		log.Fatalf("Failed to write to characteristic: %v", err)
	}

	timeout := time.NewTimer(time.Second * 10)

	select {
	case <-timeout.C:
		break
	case lines := <-respStrm.LinesCh:
		for _, l := range lines {
			log.Println("> " + l)
			if v, ok := ParseLine2Value(l); ok {
				log.Printf("# OBD resp PID: %x, DATA: %x", v.Pid, v.Data)
				return v
			}
		}
		return nil
	}
	log.Println("timeout")
	return nil
}
