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
	assert.Nil(t, v, "expected nil")

	line = "41 40 FE DC 0C 80"
	v, ok = ParseLine2Value(line)
	assert.True(t, ok, "expected true")
	assert.Equal(t, byte(0x40), v.Pid, "expected 0x40")
	assert.Equal(t, []byte{0xFE, 0xDC, 0x0C, 0x80}, v.Data, "expected [0xFE, 0xDC, 0x0C, 0x80]")
}
