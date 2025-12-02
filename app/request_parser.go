package main

import (
	"bytes"
)

type requestPacket struct {
	requestType   string
	route         string
	headers       requestHeaders
	requestMethod requestMethod
}

type requestHeaders = map[string]string

func NewRequestHeader(packet []byte) requestPacket {
	fields := bytes.Fields(packet)
	rm := NewRequestMethodFromString(string(fields[0]))
	route := fields[1]
	rh := requestHeaders{}

	for i := 3; i < len(fields); i += 2 {
		if i+1 >= len(fields) {
			break
		}

		if fields[i] == nil || fields[i+1] == nil {
			break
		}

		key := string(fields[i][:len(fields[i])-1])
		value := string(fields[i+1])
		rh[key] = value
	}

	return requestPacket{
		requestMethod: rm,
		route:         string(route),
		headers:       rh,
	}
}

func (rh requestPacket) getMethod() string {
	return rh.requestMethod.String()
}

func (rh requestPacket) getRoute() string {
	return rh.route
}
