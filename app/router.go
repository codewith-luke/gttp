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

func NewRequestMethodFromString(method MethodType) (RequestMethod, error) {
	if method == "" {
		return RequestMethod{}, fmt.Errorf("invalid method")
	}

	m := GET
	switch method {
	case ALL.key:
		m = ALL
	case GET.key:
		m = GET
	case POST.key:
		m = POST
	default:
		m = GET
	}

	return m, nil
}

type Response struct {
	StatusCode int
	Headers    RequestHeaders
}

type Router interface {
	handleRequest(conn net.Conn) error
}

type RouteHandler struct {
	handler func(headers RequestParser)
}

type RouteContext struct {
	Route       string
	Path        string
	Conn        net.Conn
	body        string
	headers     RequestHeaders
	requestType RequestMethod
	statusCode  int
}

func (c *RouteContext) Write(data []byte) {
	statusText := getStatusText(c.statusCode)
	res := fmt.Sprintf("HTTP/1.1 %d %s\r\n", c.statusCode, statusText)

	ct, ok := c.headers["Content-Type"].(string)

	if ok && len(ct) > 0 {
		res += fmt.Sprintf("Content-Type: %s\r\n", c.headers["Content-Type"])
	}

	ce, ok := c.headers["Content-Encoding"].(string)

	if ok && len(ce) > 0 {
		res += fmt.Sprintf("Content-Encoding: %s\r\n", c.headers["Content-Encoding"])
	}

	if len(data) > 0 {
		contentLength := fmt.Sprintf("Content-Length: %d", len(data))
		res += fmt.Sprintf("%s\r\n", contentLength)
	}

	res += fmt.Sprintf("\r\n%s", string(data))

	c.Conn.Write([]byte(res))
}

func (c *RouteContext) SetHeader(response Response) {
	c.statusCode = response.StatusCode
	c.headers = response.Headers
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
					context.SetHeader(Response{
						StatusCode: 404,
						Headers: RequestHeaders{
							"Content-Type": "text/plain",
						},
					})
					context.Write([]byte(""))
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

func (r RouterV2) route(conn net.Conn, requestPacket RequestParser) {
	requestedRoute := requestPacket.getRoute()

	rh := r.getHandler(conn, requestPacket, r.Routes, requestedRoute)
	rh.handler(requestPacket)
}

func (r RouterV2) parseRequest(conn net.Conn) RequestParser {
	var requestHeader = make([]byte, 1024)
	_, err := conn.Read(requestHeader)

	if err != nil {
		fmt.Println("Failed to read request requestHeader: ", err.Error())
		os.Exit(1)
	}

	return NewRequest(requestHeader)
}

func (r RouterV2) getHandler(conn net.Conn, packet RequestParser, routes Routes, requestedRoute string) RouteHandler {
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

	mh := handlers[packet.StatusLine.Method.String()]

	if mh.handler == nil || mh.method.String() != packet.StatusLine.Method.String() {
		mh = routes["/404"].handlers[ALL_METHOD]
	}

	handler := mh.handler

	return RouteHandler{
		handler: func(packet RequestParser) {
			headers := RequestHeaders{
				"Content-Type": "text/plain",
			}

			for k, v := range packet.Headers {
				headers[k] = v
			}

			c := RouteContext{
				Route:      requestedRoute,
				Path:       selectedPath,
				Conn:       conn,
				body:       packet.Body,
				statusCode: 200,
				headers:    headers,
			}

			handler(c)
		},
	}
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
