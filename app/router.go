package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
)

var (
	ALL  = requestMethod{"ALL"}
	GET  = requestMethod{"GET"}
	POST = requestMethod{"POST"}
)

func (r requestMethod) String() string {
	return r.key
}

func NewRequestMethodFromString(method string) requestMethod {
	switch method {
	case ALL.key:
		return ALL
	case GET.key:
		return GET
	case POST.key:
		return POST
	default:
		return GET
	}
}

type Router interface {
	handleRequest(conn net.Conn) error
}

type routeHandler struct {
	handler func(headers requestPacket)
}

type routeContext struct {
	route       string
	path        string
	headers     requestHeaders
	requestType requestMethod
	write       func(statusCode int, contentType string, status string, body string)
}

type requestMethod struct {
	key string
}

type routerV2 struct {
	Routes
}

type Routes = map[string]*route

type route struct {
	requestMethod requestMethod
	handler       Handler
	paths         map[string]*route
}

type Handler = func(context routeContext)

func NewRouter() routerV2 {
	r := routerV2{
		Routes: make(Routes),
	}

	r.Routes["/404"] = &route{
		handler: func(context routeContext) {
			context.write(404, "text/plain", "Not Found", "")
		},
	}

	return r
}

func (r routerV2) add(path string, handler Handler) {
	r.generateRoute(path, handler)
}

func (r routerV2) handleRequest(conn net.Conn) error {
	rh := r.parseRequest(conn)
	r.route(conn, rh)
	return nil
}

func (r routerV2) generateRoute(path string, handler Handler) {
	regRoutePath := regexp.MustCompile(`/[^/]+|/`)
	regPaths := regRoutePath.FindAllStringSubmatch(path, -1)
	paths := make([]string, 0)

	for _, p := range regPaths {
		paths = append(paths, p[0])
	}

	r.Routes = makeRoute(r.Routes, paths, handler)
}

func (r routerV2) route(conn net.Conn, requestPacket requestPacket) {
	requestedRoute := requestPacket.getRoute()

	rh := r.getHandler(conn, requestPacket, r.Routes, requestedRoute)
	rh.handler(requestPacket)
}

func (r routerV2) parseRequest(conn net.Conn) requestPacket {
	var requestHeader = make([]byte, 1024)
	_, err := conn.Read(requestHeader)

	if err != nil {
		fmt.Println("Failed to read request requestHeader: ", err.Error())
		os.Exit(1)
	}

	return NewRequestHeader(requestHeader)
}

func (r routerV2) getHandler(conn net.Conn, packet requestPacket, routes Routes, requestedRoute string) routeHandler {
	regRoutePath := regexp.MustCompile(`/[^/]+|/`)
	paths := regRoutePath.FindAllStringSubmatch(requestedRoute, -1)
	selectedPath := paths[0][0]
	selectedRoute, ok := routes[paths[0][0]]

	if !ok {
		selectedRoute = routes["/404"]
	} else {
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
				write: func(statusCode int, contentType string, status string, body string) {
					r.writeResponse(conn, statusCode, contentType, status, body)
				},
			}

			handler(c)
		},
	}
}

func (r routerV2) writeResponse(conn net.Conn, statusCode int, contentType string, status string, body string) {
	ct := fmt.Sprintf("Content-Type: %s", contentType)
	contentLength := fmt.Sprintf("Content-Length: %d", len(body))

	res := fmt.Sprintf("HTTP/1.1 %d %s\r\n%s\r\n%s\r\n\r\n%s", statusCode, status, ct, contentLength, body)
	conn.Write([]byte(res))
}

func makeRoute(routes Routes, paths []string, handler Handler) Routes {
	if handler == nil {
		return routes
	}

	if len(paths) == 0 {
		return routes
	}

	currPath := paths[0]
	newPaths := paths[1:]
	_, ok := routes[currPath]
	currRoute := routes[currPath]

	if !ok {
		currRoute = &route{
			paths: Routes{},
		}
	}

	if len(newPaths) > 0 {
		currRoute.paths = makeRoute(currRoute.paths, newPaths, handler)
		routes[currPath] = currRoute
		return routes
	}

	currRoute.handler = handler
	routes[currPath] = currRoute
	return routes
}
