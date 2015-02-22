package main

import (
	"flag"
	"fmt"
	"net"
)

var (
	port = flag.String("port", "8080", "Port to listen on")
)

func handleConnection(conn net.Conn) {
	b := []byte("Hello, world!")

	_, err := conn.Write(b)
	if err != nil {
		fmt.Println("handleConnection: ", err)
	}
}

func serve(ln net.Listener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go handleConnection(conn)
	}
}

func main() {
	flag.Parse()

	// Serve
	ln, err := net.Listen("tcp", ":"+*port)
	if err == nil {
		if err := serve(ln); err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Println(err)
	}
}
