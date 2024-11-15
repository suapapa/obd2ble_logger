package main

import (
	"fmt"
	"strings"
)

const (
	OBD2Svc01EngineSpeed  = 0x0C
	OBD2Svc01VehicleSpeed = 0x0D
)

type Value struct {
	Svc  byte
	Pid  byte
	Data []byte
}

// 41 40 FE DC 0C 80
// to
// 40 for pid
// FE DC 0C 80 for data
func ParseLine2Value(line string) (*Value, bool) {
	parts := strings.Split(line, " ")
	if len(parts) < 3 {
		return nil, false
	}

	if parts[0] != "41" {
		return nil, false
	}

	pid, err := hex2byte(parts[1])
	if err != nil {
		return nil, false
	}
	data, err := hex2bytes(parts[2:])
	if err != nil {
		return nil, false
	}

	return &Value{
		Svc:  0x01,
		Pid:  pid,
		Data: data,
	}, true
}

func (v *Value) PidString() string {
	switch v.Pid {
	case OBD2Svc01EngineSpeed:
		return "engine_speed"
	case OBD2Svc01VehicleSpeed:
		return "vehicle_speed"

	// TBD: Add more cases
	default:
		return fmt.Sprintf("%02x-%02x", v.Svc, v.Pid)
	}
}

func (v *Value) String() string {
	switch v.Pid {
	case OBD2Svc01EngineSpeed:
		return fmt.Sprint(float64(256*int(v.Data[0])+int(v.Data[1])) / 4)
	case OBD2Svc01VehicleSpeed:
		return fmt.Sprint(int(v.Data[0]))

	// TBD: Add more cases
	default:
		return fmt.Sprintf("%x", v.Data)
	}
}

func parseVal2SupportedPids(val []byte) []byte {
	var pids []byte

	currPid := byte(0x01)
	for i := 0; i < len(val); i++ {
		for b := 7; b >= 0; b-- {
			if val[i]&(1<<b) != 0 {
				pids = append(pids, currPid)
			}
			currPid++
		}
	}

	return pids
}
