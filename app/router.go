package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
)

type Router interface {
	handleRequest(conn net.Conn) error
}

type router struct {
}

type routeHandler struct {
	handler func(headers requestPacket)
}

type routeContext struct {
	route   string
	path    string
	headers requestHeaders
}

type routePath struct {
	handler func(context routeContext)
	paths   map[string]routePath
}

func NewRouter() Router {
	return &router{}
}

func (r router) handleRequest(conn net.Conn) error {
	rh := r.parseRequest(conn)
	r.route(conn, rh)
	return nil
}

func (r router) route(conn net.Conn, headers requestPacket) {
	requestedRoute := headers.getRoute()

	routes := map[string]routePath{
		"/": {
			handler: func(c routeContext) {
				r.writeResponse(conn, 200, "OK", "test")
			},
		},
		"/echo": {
			handler: func(c routeContext) {
				r.writeResponse(conn, 200, "OK", "test")
			},
			paths: map[string]routePath{
				"/:value": {
					handler: func(c routeContext) {
						res := strings.Replace(c.path, "/", "", 1)
						r.writeResponse(conn, 200, "OK", res)
					},
				},
				"/bob": {
					handler: func(c routeContext) {
						r.writeResponse(conn, 200, "OK", "1")
					},
				},
			},
		},
		"/user-agent": {
			handler: func(c routeContext) {
				userAgent := c.headers["User-Agent"]
				r.writeResponse(conn, 200, "OK", userAgent)
			},
		},
		"/404": {
			handler: func(c routeContext) {
				r.writeResponse(conn, 404, "Not Found", "")
			},
		},
	}

	rh := r.getHandler(routes, requestedRoute)
	rh.handler(headers)
}

func (r router) writeResponse(conn net.Conn, statusCode int, status string, body string) {
	contentType := "Content-Type: text/plain"
	contentLength := fmt.Sprintf("Content-Length: %d", len(body))

	res := fmt.Sprintf("HTTP/1.1 %d %s\r\n%s\r\n%s\r\n\r\n%s", statusCode, status, contentType, contentLength, body)
	conn.Write([]byte(res))
}

func (r router) parseRequest(conn net.Conn) requestPacket {
	var requestHeader = make([]byte, 1024)
	_, err := conn.Read(requestHeader)

	if err != nil {
		fmt.Println("Failed to read request requestHeader: ", err.Error())
		os.Exit(1)
	}

	return NewRequestHeader(requestHeader)
}

func (r router) getHandler(routes map[string]routePath, requestedRoute string) routeHandler {
	regRoutePath := regexp.MustCompile(`/[^/]+|/`)
	paths := regRoutePath.FindAllStringSubmatch(requestedRoute, -1)
	selectedPath := paths[0][0]
	selectedRoute := routes[paths[0][0]]

	for i := 1; i < len(paths); i++ {
		selectedPath = paths[i][0]
		r := selectedRoute.paths[paths[i][0]]

		if r.handler != nil {
			selectedRoute = r
		} else if selectedRoute.paths["/:value"].handler != nil {
			selectedRoute = selectedRoute.paths["/:value"]
		} else {
			selectedRoute = routes["/404"]
			selectedPath = "/404"
		}
	}

	handler := selectedRoute.handler

	if handler == nil {
		handler = routes["/404"].handler
	}

	return routeHandler{
		handler: func(packet requestPacket) {
			c := routeContext{
				route:   requestedRoute,
				path:    selectedPath,
				headers: packet.headers,
			}

			handler(c)
		},
	}
}
