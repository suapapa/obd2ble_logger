package main

import (
	"strings"
)

type RespStream struct {
	buff    []byte
	LinesCh chan []string
}

func NewRespStream() *RespStream {
	return &RespStream{
		LinesCh: make(chan []string),
	}
}

func (r *RespStream) Reset() {
	r.buff = []byte{}
}

func (r *RespStream) Write(p []byte) {
	r.buff = append(r.buff, p...)
	if r.buff[len(r.buff)-1] == '>' {
		r.buff = r.buff[:len(r.buff)-1]
		rLines := strings.Split(string(r.buff), "\r")
		var lines []string
		for _, l := range rLines {
			if l != "" {
				lines = append(lines, strings.TrimSpace(l))
			}
		}
		r.LinesCh <- lines

		r.Reset()
	}
}
