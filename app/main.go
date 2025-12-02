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

	router.add("/", func(context routeContext) {
		context.write(200, "text/plain", "OK", "test")
	})

	router.add("/files/:value", func(context routeContext) {
		res := context.path[1:]
		fileContent, err := getFileContent(res)

		if err != nil {
			context.write(404, "text/plain", "Not Found", "")
			return
		}

		context.write(200, "application/octet-stream", "OK", string(fileContent))
	})

	router.add("/echo", func(context routeContext) {
		context.write(200, "text/plain", "OK", "test")
	})

	router.add("/echo/:value", func(context routeContext) {
		res := context.path[1:]
		context.write(200, "text/plain", "OK", res)
	})

	router.add("/user-agent", func(context routeContext) {
		userAgent := context.headers["User-Agent"]
		context.write(200, "text/plain", "OK", userAgent)
	})

	router.add("/hello", func(context routeContext) {
		context.write(200, "text/plain", "OK", "Hello World!")
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
