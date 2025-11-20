package main

import "bytes"

type RequestHeader struct {
	requestType string
	route       string
}

func NewRequestHeader(packet []byte) RequestHeader {
	fields := bytes.Fields(packet)
	requestType := fields[0]
	route := fields[1]

	return RequestHeader{
		requestType: string(requestType),
		route:       string(route),
	}
}

func (rh RequestHeader) getType() string {
	return rh.requestType
}

func (rh RequestHeader) getRoute() string {
	return rh.route
}
