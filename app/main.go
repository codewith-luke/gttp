package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"os"
	"slices"
	"strings"
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
		context.Write([]byte("Hello World!"))
	})

	router.add(GET, "/files/:value", func(context RouteContext) {
		res := context.Path[1:]
		fileContent, err := getFileContent(res)

		if err != nil {
			context.SetHeader(Response{
				StatusCode: 404,
			})
			context.Write([]byte(""))
			return
		}

		context.Write(fileContent)
	})

	router.add(POST, "/files/:value", func(context RouteContext) {
		res := context.Path[1:]
		ok, err := createFile(res, context.body)

		if err != nil {
			context.SetHeader(Response{
				StatusCode: 404,
			})
			context.Write([]byte(""))
			return
		}

		if !ok {
			context.SetHeader(Response{
				StatusCode: 500,
			})
			context.Write([]byte(""))
		}

		context.SetHeader(Response{
			StatusCode: 201,
			Headers: RequestHeaders{
				"Content-Type": "application/octet-stream",
			},
		})
		context.Write([]byte(""))
	})

	router.add(GET, "/echo", func(context RouteContext) {
		context.Write([]byte("test"))
	})

	router.add(GET, "/echo/:value", func(context RouteContext) {
		acceptEncoding, ok := context.headers["Accept-Encoding"]
		val := context.Path[1:]
		headers := context.headers

		if ok {
			encodings, ok := acceptEncoding.([]string)

			if ok && slices.Contains(encodings, "gzip") {
				headers["Content-Encoding"] = "gzip"
				context.SetHeader(Response{
					StatusCode: 200,
					Headers:    headers,
				})
			}

			r := strings.NewReader(val)
			w := gzip.NewWriter(&context)
			defer w.Close()
			io.Copy(w, r)
			return
		}

		context.Write([]byte(val))
	})

	router.add(GET, "/user-agent", func(context RouteContext) {
		userAgent := context.headers["User-Agent"].(string)
		context.Write([]byte(userAgent))
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
