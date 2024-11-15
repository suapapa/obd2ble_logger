package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-ble/ble"
	"github.com/pkg/errors"
)

var (
	// Target device name or address
	deviceName = "IOS-Vlink"
	// deviceAddr = "d2:e0:2f:8d:49:55"

	// UUIDs for service and characteristic
	serviceUUID        = ble.MustParse("E7810A71-73AE-499D-8C15-FAA9AEF0C3F2")
	characteristicUUID = ble.MustParse("BEF8D6C9-9C21-4C9E-B632-BD58C1009F9F")
)

type ELM327BLE struct {
	c        ble.Client
	char     *ble.Characteristic
	respStrm *RespStream

	PIDs []byte // Supported PIDs

	enableLogging bool
}

func NewELM327BLE(debug bool) (*ELM327BLE, error) {
	p := &ELM327BLE{
		enableLogging: debug,
	}
	// Scan for the target device
	p.debug("Scanning for BLE devices...")
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), 10*time.Second))
	client, err := ble.Connect(ctx, func(a ble.Advertisement) bool {
		if a.LocalName() == deviceName {
			p.debug("Found target device: %s(%s)", deviceName, a.Addr())
			return true
		}
		return false
	})
	if err != nil {
		p.debug("Failed to connect to device: %v", err)
		return nil, errors.Wrap(err, "failed to connect to device")
	}

	// Discover the specified service
	p.debug("Discovering services...")
	profile, err := client.DiscoverProfile(true)
	if err != nil {
		p.Close()
		return nil, errors.Wrap(err, "failed to discover profile")
	}

	var tChar *ble.Characteristic
	// var targetCharacteristic *ble.Characteristic
	for _, service := range profile.Services {
		if service.UUID.Equal(serviceUUID) {
			log.Println("Found target service")
			for _, char := range service.Characteristics {
				if char.UUID.Equal(characteristicUUID) {
					log.Println("Found target characteristic")
					tChar = char
					break
				}
			}
		}
	}

	if tChar == nil {
		p.Close()
		return nil, errors.New("target characteristic not found")
	}

	p.c = client
	p.char = tChar
	p.respStrm = NewRespStream()

	err = client.Subscribe(tChar, false, func(data []byte) {
		// log.Printf("notify: %x, %s", data, string(data))
		p.respStrm.Write(data)
	})
	if err != nil {
		p.Close()
		return nil, errors.Wrap(err, "failed to subscribe to characteristic")
	}

	if err = p.init(); err != nil {
		p.Close()
		return nil, errors.Wrap(err, "failed to init")
	}

	if err = p.querySupportedPids(); err != nil {
		p.Close()
		return nil, errors.Wrap(err, "failed to get supported pids")
	}

	p.debug("Initialized")

	return p, nil
}

func (e *ELM327BLE) Close() error {
	return e.c.CancelConnection()
}

func (e *ELM327BLE) ReadCurrData(pid byte) (*Value, error) {
	if !e.isSupportedPID(pid) {
		return nil, errors.Errorf("PID %02x is not supported", pid)
	}

	cmd := fmt.Sprintf("01 %02x", pid) // Service 01 - Show current data
	return e.sendCmd(cmd)
}

func (e *ELM327BLE) ReadCurrDatas(pids []byte) ([]*Value, error) {
	var vs []*Value
	for _, pid := range pids {
		v, err := e.ReadCurrData(pid)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read PID %02x", pid)
		}
		vs = append(vs, v)
	}
	return vs, nil
}

func (e *ELM327BLE) isSupportedPID(pid byte) bool {
	var found bool
	for _, p := range e.PIDs {
		if p == pid {
			found = true
			break
		}
	}
	return found
}

func (e *ELM327BLE) querySupportedPids() error {
	v, err := e.sendCmd("01 00")
	if err != nil {
		return errors.Wrap(err, "failed to get supported pids")
	}

	e.PIDs = parseVal2SupportedPids(v.Data)
	return nil
}

func (e *ELM327BLE) init() error {
	commands := []string{
		"ATZ", // ELM327 v2.3
		"AT E0",
		"AT L0",
		"AT SP 00",
		// "01 00",
	}
	for _, cmd := range commands {
		e.sendCmd(cmd)
		// if _, err := e.sendCmd(cmd); err != nil {
		// 	return errors.Wrap(err, "failed to init")
		// }
		e.sendCmd(cmd)
	}
	return nil
}

func (e *ELM327BLE) sendCmd(command string) (*Value, error) {
	e.debug("Writing command: %s", command)

	cmd := append([]byte(command), '\r')
	if err := e.c.WriteCharacteristic(e.char, cmd, false); err != nil {
		e.debug("Failed to write to characteristic: %v", err)
		return nil, errors.Wrap(err, "failed to write to characteristic")
	}

	timeout := time.NewTimer(10 * time.Second)

	select {
	case <-timeout.C:
		break
	case lines := <-e.respStrm.LinesCh:
		for _, l := range lines {
			e.debug("> " + l)
			if v, ok := ParseLine2Value(l); ok {
				// log.Printf("# OBD resp PID: %x, DATA: %x", v.Pid, v.Data)
				cs := strings.Split(command, " ")
				if len(cs) < 2 {
					e.debug("Invalid command: %s", command)
					continue
				}

				hexPid, _ := hex2byte(cs[1])
				if len(cs) > 1 && hexPid != v.Pid {
					e.debug("PID mismatch: %02x != %02x", hexPid, v.Pid)
					continue
				}

				return v, nil
			}
		}
		return nil, errors.New("resp has no valid value")
	}
	e.debug("timeout")
	return nil, errors.New("timeout")
}

func (e *ELM327BLE) debug(format string, v ...interface{}) {
	if e.enableLogging {
		log.Printf(format, v...)
	}
}
