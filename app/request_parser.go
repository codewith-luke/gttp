package main

import (
	"bytes"
	"fmt"
	"strconv"
)

type RequestPacket struct {
	Route   string
	Headers requestHeaders
	Method  RequestMethod
	Body    string
}

type requestHeaders = map[string]any

func NewRequestHeader(packet []byte) RequestPacket {
	fields := bytes.Fields(packet)
	rm := NewRequestMethodFromString(MethodType(fields[0]))
	route := string(fields[1])
	rh := requestHeaders{}
	newLineIndex := bytes.Index(packet, []byte("\r\n\r\n")) + 4
	requestHead := packet[:newLineIndex]
	fieldsB := bytes.Fields(requestHead)
	fmt.Println(fieldsB)

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

	body := ""
	cl, ok := rh["Content-Length"].(int)

	if !ok {
		cl = 0
	}

	if cl == 0 {
		return RequestPacket{
			Method:  rm,
			Route:   route,
			Headers: rh,
		}
	}

	if newLineIndex != -1 {
		body = string(packet[newLineIndex : newLineIndex+cl])
	}

	return RequestPacket{
		Method:  rm,
		Route:   route,
		Headers: rh,
		Body:    body,
	}
}

func (rh RequestPacket) getMethod() MethodType {
	return rh.Method.String()
}

func (rh RequestPacket) getRoute() string {
	return rh.Route
}
