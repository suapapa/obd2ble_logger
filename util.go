package main

import (
	"fmt"
)

func hex2byte(s string) (byte, error) {
	var n byte
	for i := 0; i < len(s); i++ {
		n <<= 4
		switch {
		case '0' <= s[i] && s[i] <= '9':
			n |= s[i] - '0'
		case 'a' <= s[i] && s[i] <= 'f':
			n |= s[i] - 'a' + 10
		case 'A' <= s[i] && s[i] <= 'F':
			n |= s[i] - 'A' + 10
		default:
			return 0, fmt.Errorf("invalid hex character: %q", s[i])
		}
	}
	return n, nil
}

func hex2bytes(ss []string) ([]byte, error) {
	var b []byte
	for _, s := range ss {
		n, err := hex2byte(s)
		if err != nil {
			return nil, err
		}
		b = append(b, n)
	}
	return b, nil
}
