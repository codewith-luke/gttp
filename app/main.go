package main

import (
	"fmt"
	"net"
	"os"
)

var _ = net.Listen
var _ = os.Exit
var args appArgs

func main() {
	args = parseArguments(os.Args)
	createDirectory(args.directory)
	fmt.Println("Logs from your program will appear here!")

	ln, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	defer ln.Close()

	router := NewRouter()

	router.add(GET, "/", func(context RouteContext) {
		context.write(Response{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type": "text/plain",
			},
			Body: "test",
		})
	})

	router.add(GET, "/files/:value", func(context RouteContext) {
		res := context.path[1:]
		fileContent, err := getFileContent(res)

		if err != nil {
			context.write(Response{
				StatusCode: 404,
				Headers: map[string]string{
					"Content-Type": "text/plain",
				},
			})
			return
		}

		context.write(Response{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type": "application/octet-stream",
			},
			Body: string(fileContent),
		})
	})

	router.add(POST, "/files/:value", func(context RouteContext) {
		res := context.path[1:]
		ok, err := createFile(res, context.body)

		if err != nil {
			context.write(Response{
				StatusCode: 404,
				Headers: map[string]string{
					"Content-Type": "text/plain",
				},
			})
			return
		}

		if !ok {
			context.write(Response{
				StatusCode: 500,
				Headers: map[string]string{
					"Content-Type": "text/plain",
				},
			})
		}

		context.write(Response{
			StatusCode: 201,
			Headers: map[string]string{
				"Content-Type": "application/octet-stream",
			},
		})
	})

	router.add(GET, "/test/me", func(context RouteContext) {
		context.write(Response{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type": "text/plain",
			},
			Body: "test",
		})
	})

	router.add(GET, "/echo", func(context RouteContext) {
		context.write(Response{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type": "text/plain",
			},
			Body: "test",
		})
	})

	router.add(GET, "/echo/:value", func(context RouteContext) {
		acceptEncoding := context.headers["Accept-Encoding"].(string)
		res := context.path[1:]

		if acceptEncoding == "gzip" {
			context.write(Response{
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type":     "text/plain",
					"Content-Encoding": "gzip",
				},
				Body: res,
			})
		} else {
			context.write(Response{
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type": "text/plain",
				},
				Body: res,
			})
		}
	})

	router.add(GET, "/user-agent", func(context RouteContext) {
		userAgent := context.headers["User-Agent"].(string)
		context.write(Response{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type": "text/plain",
			},
			Body: userAgent,
		})
	})

	router.add(GET, "/hello", func(context RouteContext) {
		context.write(Response{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type": "text/plain",
			},
			Body: "Hello World!",
		})
	})

	for {
		conn, err := ln.Accept()

		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn, router)
	}
}

func handleConnection(conn net.Conn, router Router) {
	fmt.Println("Logs from your program will appear here!")
	err := router.handleRequest(conn)

	if err != nil {
		fmt.Println("Error handling request: ", err.Error())
		os.Exit(1)
	}

	conn.Close()
}
