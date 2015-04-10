package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/afking/godelta/delta"
	"github.com/golang/protobuf/proto"

	"github.com/codegangsta/cli"
)

const (
	TIMEOUT time.Duration = time.Second * 4

	CONN_HOST string = "192.168.1.10"
	CONN_PORT string = "80" //"2616"
	CONN_TYPE string = "tcp"
)

var (
	conn *net.TCPConn
	add  *net.TCPAddr
	err  error
)

func init() {
	add, err = net.ResolveTCPAddr(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		log.Fatal(err)
	}

	status := make(chan bool, 1)
	go func() {
		conn, err = net.DialTCP(CONN_TYPE, nil, add)
		if err != nil {
			log.Fatal(err)
		}
		status <- true
	}()

	go func() {
		time.Sleep(TIMEOUT)
		status <- false
	}()

	if state := <-status; !state {
		log.Fatal(fmt.Errorf("TCP timeout"))
	}

	if err := conn.SetKeepAlive(true); err != nil {
		log.Fatal(err)
	}
	if err := conn.SetKeepAlivePeriod(TIMEOUT); err != nil {
		log.Fatal(err)
	}
}

func read() (*delta.Message, error) {
	b := make([]byte, 128)
	if err := conn.SetReadDeadline(time.Now().Add(TIMEOUT)); err != nil {
		return nil, err
	}
	if _, err := conn.Read(b); err != nil {
		return nil, err
	}

	msg := &delta.Message{}
	if err = proto.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func write(msg *delta.Message) error {
	b, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	if _, err := conn.Write(b); err != nil {
		return err
	}
	return nil
}

func msgType(t delta.Message_Type) error {
	msg := &delta.Message{
		Type: t.Enum(),
	}

	return write(msg)
}

func msgPoint(x, y, z float64) error {
	msg := &delta.Message{
		Type: delta.Message_POINT.Enum(),
		Point: &delta.Point{
			X: &x,
			Y: &y,
			Z: &z,
		},
	}

	if err := write(msg); err != nil {
		return err
	}
	return nil
}

// e wraps errors for application commands
func e(f func(*cli.Context) error) func(*cli.Context) {
	return func(c *cli.Context) {
		if err := f(c); err != nil {
			log.Println("error: ", err)
			return
		}
	}
}

// ping delta arm robot
func ping(c *cli.Context) error {
	msg := &delta.Message{
		Type: delta.Message_PING.Enum(),
	}

	startTime := time.Now()
	if err = write(msg); err != nil {
		return err
	}
	rsp, err := read()
	if err != nil {
		return err
	}
	endTime := time.Now()

	if rsp.Type == msg.Type {
		fmt.Println("pong [%v]", endTime.Sub(startTime))
	} else {
		return fmt.Errorf("expected ping got %i", rsp.Type)
	}
	return nil
}

func xbox(c *cli.Context) error {
	return xboxDriver()
}

func main() {
	app := cli.NewApp()
	app.Name = "delta"
	app.Usage = "Delta arm go client, <3 Ed"
	app.Action = func(c *cli.Context) {
		fmt.Println("Go Delta Arm Client")
	}
	app.Commands = []cli.Command{
		{
			Name:    "ping",
			Aliases: []string{"p"},
			Usage:   "ping message to delta arm robot",
			Action:  e(ping),
		},
		{
			Name:    "xbox",
			Aliases: []string{"x"},
			Usage:   "xbox control",
			Action:  e(xbox),
		},
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Panic ", r)
		}

		if conn != nil {
			conn.Close()
		}
	}()

	app.Run(os.Args)
}
