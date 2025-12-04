package main

import (
	"bytes"
	"strconv"
)

type requestPacket struct {
	route   string
	headers requestHeaders
	method  requestMethod
	body    string
}

type requestHeaders = map[string]any

func NewRequestHeader(packet []byte) requestPacket {
	fields := bytes.Fields(packet)
	rm := NewRequestMethodFromString(MethodType(fields[0]))
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
		field := fields[i+1]

		var value any
		val, err := strconv.Atoi(string(field))

		if err == nil {
			value = val
		} else {
			value = string(fields[i+1])
		}

		rh[key] = value
	}

	newLineIndex := bytes.Index(packet, []byte("\r\n\r\n")) + 4
	body := ""
	cl, ok := rh["Content-Length"].(int)

	if !ok {
		cl = 0
	}

	if cl == 0 {
		return requestPacket{
			method:  rm,
			route:   string(route),
			headers: rh,
		}
	}

	if newLineIndex != -1 {
		body = string(packet[newLineIndex : newLineIndex+cl])
	}

	return requestPacket{
		method:  rm,
		route:   string(route),
		headers: rh,
		body:    body,
	}
}

func (rh requestPacket) getMethod() MethodType {
	return rh.method.String()
}

func (rh requestPacket) getRoute() string {
	return rh.route
}
