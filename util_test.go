package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse2Value(t *testing.T) {
	// line := "41 40 FE DC 0C 80"
	line := "ELM327 v2.3"
	v, ok := ParseLine2Value(line)
	assert.False(t, ok, "expected false")

	line = "41 40 FE DC 0C 80"
	v, ok = ParseLine2Value(line)
	assert.True(t, ok, "expected true")
	assert.Equal(t, v.Pid, byte(0x40), "expected 0x40")
	assert.Equal(t, v.Data, []byte{0xFE, 0xDC, 0x0C, 0x80}, "expected [0xFE, 0xDC, 0x0C, 0x80]")

}
