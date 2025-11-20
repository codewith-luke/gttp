package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
)

type Router interface {
	handleRequest(conn net.Conn) error
}

type router struct {
}

func NewRouter() Router {
	return &router{}
}

func (r router) handleRequest(conn net.Conn) error {
	rh := r.parseRequest(conn)
	r.route(conn, rh)
	return nil
}

func (r router) route(conn net.Conn, header RequestHeader) {
	route := header.getRoute()

	reg, _ := regexp.Compile("^/(.+)/(.+)")

	subMatch := reg.FindAllStringSubmatch(route, -1)
	fullRoute := subMatch[0]
	path := fmt.Sprintf("/%s", fullRoute[1])

	switch path {
	case "/":
		r.writeResponse(conn, 200, "OK", "test")
		break
	case "/echo":
		subPath := fullRoute[2]
		r.writeResponse(conn, 200, "OK", subPath)
		break
	default:
		r.writeResponse(conn, 404, "Not Found", "")
		break
	}
}

func (r router) writeResponse(conn net.Conn, statusCode int, status string, body string) {
	contentType := "Content-Type: text/plain"
	contentLength := fmt.Sprintf("Content-Length: %d", len(body))

	res := fmt.Sprintf("HTTP/1.1 %d %s\r\n%s\r\n%s\r\n%s\r\n", statusCode, status, contentType, contentLength, body)
	conn.Write([]byte(res))
}

func (r router) parseRequest(conn net.Conn) RequestHeader {
	var requestHeader = make([]byte, 1024)
	_, err := conn.Read(requestHeader)

	if err != nil {
		fmt.Println("Failed to read request requestHeader: ", err.Error())
		os.Exit(1)
	}

	return NewRequestHeader(requestHeader)
}
