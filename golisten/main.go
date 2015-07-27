package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

const (
	CONN_HOST string = "192.168.1.100"
	CONN_PORT string = "2616"
	CONN_TYPE string = "tcp"
)

func main() {
	if l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT); err != nil {

		log.Fatal(err)
	}
	defer l.Close()

	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	var buf bytes.Buffer
	n, err := io.Copy(&buf, conn)
	if err != nil {
		fmt.Println(err)
	}

	log.Printf("%n bytes received\n", f)
	conn.Close()
}
