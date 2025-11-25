package main

import (
	"bytes"
	"strings"
)

type requestPacket struct {
	requestType string
	route       string
	headers     requestHeaders
}

type requestHeaders = map[string]string

func NewRequestHeader(packet []byte) requestPacket {
	fields := bytes.Fields(packet)
	requestType := fields[0]
	route := fields[1]
	rh := requestHeaders{}

	for i := 3; i < len(fields); i += 2 {
		if i+1 >= len(fields) {
			break
		}

		if fields[i] == nil || fields[i+1] == nil {
			break
		}

		key := strings.Replace(string(fields[i]), ":", "", 1)
		value := string(fields[i+1])
		rh[key] = value
	}

	return requestPacket{
		requestType: string(requestType),
		route:       string(route),
		headers:     rh,
	}
}

func (rh requestPacket) getType() string {
	return rh.requestType
}

func (rh requestPacket) getRoute() string {
	return rh.route
}
