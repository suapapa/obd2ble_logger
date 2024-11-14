package main

import "strings"

type Value struct {
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
		Pid:  pid,
		Data: data,
	}, true
}
