package main

import (
	"fmt"
	"net"
	"os"
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

	switch route {
	case "/":
		r.writeResponse(conn, 200, "OK")
		break
	default:
		r.writeResponse(conn, 404, "Not Found")
		break

	}
}

func (r router) writeResponse(conn net.Conn, status int, reason string) {
	res := fmt.Sprintf("HTTP/1.1 %d %s\r\n\r\n", status, reason)
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
