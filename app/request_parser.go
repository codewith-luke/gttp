package main

import "bytes"

type RequestHeaders struct {
	requestType string
	route       string
	userAgent   string
}

func NewRequestHeader(packet []byte) RequestHeaders {
	fields := bytes.Fields(packet)
	requestType := fields[0]
	route := fields[1]
	userAgent := fields[8]

	return RequestHeaders{
		requestType: string(requestType),
		userAgent:   string(userAgent),
		route:       string(route),
	}
}

func (rh RequestHeaders) getType() string {
	return rh.requestType
}

func (rh RequestHeaders) getRoute() string {
	return rh.route
}
