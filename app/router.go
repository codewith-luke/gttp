package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
)

type MethodType string

const (
	GET_METHOD  MethodType = "GET"
	POST_METHOD MethodType = "POST"
	ALL_METHOD  MethodType = "ALL"
)

var (
	ALL  = RequestMethod{ALL_METHOD}
	GET  = RequestMethod{GET_METHOD}
	POST = RequestMethod{POST_METHOD}
)

type RequestMethod struct {
	key MethodType
}

func (r RequestMethod) String() MethodType {
	return r.key
}

func NewRequestMethodFromString(method MethodType) RequestMethod {
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

type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       string
}

type WriteResponse struct {
	Conn       net.Conn
	StatusText string
	Response
}

type Router interface {
	handleRequest(conn net.Conn) error
}

type RouteHandler struct {
	handler func(headers RequestPacket)
}

type RouteContext struct {
	route       string
	path        string
	headers     requestHeaders
	requestType RequestMethod
	write       func(response Response)
	body        string
}

type RouterV2 struct {
	Routes
}

type Routes = map[string]*Route

type Route struct {
	handlers map[MethodType]MethodHandler
	paths    map[string]*Route
}

type Handler = func(context RouteContext)
type MethodHandler struct {
	method  RequestMethod
	handler Handler
}

func NewRouter() RouterV2 {
	r := RouterV2{
		Routes: make(Routes),
	}

	r.Routes["/404"] = &Route{
		handlers: map[MethodType]MethodHandler{
			ALL_METHOD: {
				method: ALL,
				handler: func(context RouteContext) {
					context.write(Response{
						StatusCode: 404,
						Headers: map[string]string{
							"Content-Type": "text/plain",
						},
					})
				},
			},
		},
	}

	return r
}

func (r RouterV2) add(method RequestMethod, path string, handler Handler) {
	r.generateRoute(method, path, handler)
}

func (r RouterV2) handleRequest(conn net.Conn) error {
	rh := r.parseRequest(conn)
	r.route(conn, rh)
	return nil
}

func (r RouterV2) generateRoute(method RequestMethod, path string, handler Handler) {
	regRoutePath := regexp.MustCompile(`/[^/]+|/`)
	regPaths := regRoutePath.FindAllStringSubmatch(path, -1)
	paths := make([]string, 0)

	for _, p := range regPaths {
		paths = append(paths, p[0])
	}

	r.Routes = makeRoute(method, r.Routes, paths, handler)
}

func (r RouterV2) route(conn net.Conn, requestPacket RequestPacket) {
	requestedRoute := requestPacket.getRoute()

	rh := r.getHandler(conn, requestPacket, r.Routes, requestedRoute)
	rh.handler(requestPacket)
}

func (r RouterV2) parseRequest(conn net.Conn) RequestPacket {
	var requestHeader = make([]byte, 1024)
	_, err := conn.Read(requestHeader)

	if err != nil {
		fmt.Println("Failed to read request requestHeader: ", err.Error())
		os.Exit(1)
	}

	return NewRequestHeader(requestHeader)
}

func (r RouterV2) getHandler(conn net.Conn, packet RequestPacket, routes Routes, requestedRoute string) RouteHandler {
	regRoutePath := regexp.MustCompile(`/[^/]+|/`)
	paths := regRoutePath.FindAllStringSubmatch(requestedRoute, -1)
	selectedPath := paths[0][0]
	selectedRoute, ok := routes[paths[0][0]]

	if !ok {
		selectedRoute = routes["/404"]
	} else {
		for i := 1; i < len(paths); i++ {
			selectedPath = paths[i][0]
			r, ok := selectedRoute.paths[paths[i][0]]

			if ok {
				selectedRoute = r
			}

			v, ok := selectedRoute.paths["/:value"]

			if ok && v.handlers != nil {
				selectedRoute = v
			}

			if selectedRoute.handlers == nil {
				selectedRoute = routes["/404"]
				selectedPath = "/404"
			}
		}
	}

	handlers := selectedRoute.handlers

	if handlers == nil {
		handlers = routes["/404"].handlers
	}

	mh := handlers[packet.Method.String()]

	if mh.handler == nil || mh.method.String() != packet.Method.String() {
		mh = routes["/404"].handlers[ALL_METHOD]
	}

	handler := mh.handler

	return RouteHandler{
		handler: func(packet RequestPacket) {
			c := RouteContext{
				route:   requestedRoute,
				path:    selectedPath,
				headers: packet.Headers,
				body:    packet.Body,
				write: func(response Response) {
					r.writeResponse(WriteResponse{
						Conn:       conn,
						Response:   response,
						StatusText: getStatusText(response.StatusCode),
					})
				},
			}

			handler(c)
		},
	}
}

func (r RouterV2) writeResponse(wr WriteResponse) {
	res := fmt.Sprintf("HTTP/1.1 %d %s\r\n", wr.StatusCode, wr.StatusText)

	if len(wr.Headers["Content-Type"]) > 0 {
		res += fmt.Sprintf("Content-Type: %s\r\n", wr.Headers["Content-Type"])
	}

	if len(wr.Headers["Content-Encoding"]) > 0 {
		res += fmt.Sprintf("Content-Encoding: %s\r\n", wr.Headers["Content-Encoding"])
	}

	if len(wr.Body) > 0 {
		contentLength := fmt.Sprintf("Content-Length: %d", len(wr.Body))
		res += fmt.Sprintf("%s\r\n", contentLength)
	}

	res += fmt.Sprintf("\r\n%s", wr.Body)

	wr.Conn.Write([]byte(res))
}

func getStatusText(code int) string {
	status := ""

	if code < 400 {
		switch code {
		case 201:
			status = "Created"
			break
		default:
			status = "OK"
		}
	} else {
		switch code {
		case 404:
			status = "Not Found"
			break
		default:
			status = "Internal Server Error"
		}
	}

	return status
}

func makeRoute(method RequestMethod, routes Routes, paths []string, handler Handler) Routes {
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
		currRoute = &Route{
			paths: Routes{},
		}
	}

	if len(newPaths) > 0 {
		currRoute.paths = makeRoute(method, currRoute.paths, newPaths, handler)
		routes[currPath] = currRoute
		return routes
	}

	if currRoute.handlers == nil {
		currRoute.handlers = map[MethodType]MethodHandler{}
	}

	mh := MethodHandler{method, handler}
	currRoute.handlers[mh.method.String()] = mh
	routes[currPath] = currRoute
	return routes
}
