package main

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type RequestParser struct {
	Body       string
	Headers    RequestHeaders
	StatusLine RequestStatusLine
}

type RequestPacket struct {
	RequestStatusLine
	RequestHeaders
	RequestBody
}

type RequestStatusLine struct {
	Method  RequestMethod
	Route   string
	Version string
}

type RequestBody struct {
	Body string
}

type RequestHeaders = map[string]any

func NewRequest(packet []byte) RequestParser {
	bodyIndex := bytes.Index(packet, []byte("\r\n\r\n")) + 4
	headerCollection := bytes.Split(packet[:bodyIndex-4], []byte("\r\n"))

	rp := RequestParser{}

	rp.parseStatusLine(headerCollection[:1])
	rp.parseRequestHeaders(headerCollection[1:])
	rp.parseRequestBody(packet[bodyIndex:])

	return rp
}

func (rp *RequestParser) getMethod() RequestMethod {
	return rp.StatusLine.Method
}

func (rp *RequestParser) getRoute() string {
	return rp.StatusLine.Route
}

func (rp *RequestParser) parseStatusLine(data [][]byte) {
	values := bytes.Split(data[0], []byte(" "))
	method, err := NewRequestMethodFromString(MethodType(values[0]))

	if err != nil {
		fmt.Println("invalid request method", err)
		os.Exit(1)
	}

	route := string(values[1])

	if len(route) == 0 {
		fmt.Println("invalid request Route")
		os.Exit(1)
	}

	version := string(values[2])

	if len(version) == 0 {
		fmt.Println("invalid version")
		os.Exit(1)
	}

	rp.StatusLine = RequestStatusLine{
		Method:  method,
		Route:   route,
		Version: version,
	}
}

func (rp *RequestParser) parseRequestHeaders(data [][]byte) {
	headers := RequestHeaders{}

	for _, h := range data {
		values := bytes.Split(h, []byte(":"))
		key := string(values[0])[:len(values[0])]
		value := string(values[1])

		if key == "Accept-Encoding" {
			d := strings.Split(value, ",")
			for i := range d {
				d[i] = strings.TrimSpace(d[i])
			}
			headers[key] = d
		} else {
			fieldValue := strings.TrimSpace(value)
			val, err := strconv.Atoi(fieldValue)

			if err == nil {
				headers[key] = val
			} else {
				headers[key] = fieldValue
			}

		}
	}

	rp.Headers = headers
}

func (rp *RequestParser) parseRequestBody(data []byte) {
	body := ""
	cl, ok := rp.Headers["Content-Length"].(int)

	if !ok {
		cl = 0
	}

	body = string(data[:cl])
	rp.Body = body
}

func (rp *RequestParser) getRequest() RequestPacket {
	return RequestPacket{
		RequestStatusLine: RequestStatusLine{},
		RequestHeaders:    nil,
		RequestBody:       RequestBody{},
	}
}
