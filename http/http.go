package http

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/amitsuthar69/libs/http/request"
)

// Creates a TCP listener, accepts incoming connection and parses the Request.
func Run() error {
	ln, err := net.Listen("tcp", ":3000")
	if err != nil {
		fmt.Println("Err: ", err)
	}
	defer ln.Close()
	fmt.Printf("Started tcp server: %v\n", time.Now().UTC().Local())

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Err: ", err)
			continue
		}
		fmt.Printf("New client connected: %v at %v \n", conn.RemoteAddr(), time.Now().UTC().Local())

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	req, err := request.RequestFromReader(conn)
	if err != nil {
		if err == io.EOF {
			fmt.Println("Client disconnected:", conn.RemoteAddr())
			return
		}
		fmt.Println("Request parsing error:", err)
		return
	}

	// for now, we just print the headers
	fmt.Printf("Parsed Request:\nMethod: %s\nPath: %s\nVersion: %s\n",
		req.RequestLine.Method, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion)
}
