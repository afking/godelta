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

func TCP() {
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

func read(msg *delta.Message) error {
	data := make([]byte, 128)
	//if err := conn.SetReadDeadline(time.Now().Add(TIMEOUT)); err != nil {
	//	return nil, err
	//}
	n, err := conn.Read(data)
	if err != nil {
		return err
	}
	log.Printf("%i bytes read\n", n)
	return proto.Unmarshal(data, msg)
}

func write(msg *delta.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	n, err := conn.Write(data)
	if err != nil {
		return err
	}
	log.Printf("%i bytes written\n", n)
	return nil
}

func msgType(t delta.Message_Type) error {
	msg := &delta.Message{
		Type: &t,
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

// e wraps errors for TCP application commands
func e(f func(*cli.Context) error) func(*cli.Context) {
	return func(c *cli.Context) {
		TCP() // Setup
		if err := f(c); err != nil {
			log.Println("error: ", err)
			return
		}
	}
}

// ping delta arm robot
func ping(c *cli.Context) error {
	ping := delta.Message_PING
	msg := &delta.Message{
		Type: &ping,
	}
	fmt.Println("Struct type: ", msg.GetType().String())

	startTime := time.Now()
	if err = write(msg); err != nil {
		return err
	}
	rsp := &delta.Message{}
	if err := read(rsp); err != nil {
		return err
	}
	endTime := time.Now()

	if rsp.GetType() == msg.GetType() {
		fmt.Println("pong [%v]", endTime.Sub(startTime))
	} else {
		return fmt.Errorf("Invalid type received %i", rsp.GetType().String())
	}
	return nil
}

func xbox(c *cli.Context) error {
	return xboxDriver()
}

func listen(c *cli.Context) error {
	msg := &delta.Message{}
	for {
		if err := read(msg); err != nil {
			log.Println("Listen: Error: ", err)
		} else {
			log.Println("Listen: Got type: ", msg.GetType().String())
		}
	}
}

func test(c *cli.Context) {
	p := delta.Message_PING
	msg := &delta.Message{
		Type: &p,
		Info: proto.String("Hello, world!"),
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		log.Fatal(err)
	}

	rsp := &delta.Message{}
	if err := proto.Unmarshal(data, rsp); err != nil {
		log.Fatal(err)
	}

	if msg.GetType() != rsp.GetType() {
		log.Fatalf("Data mismatch %q != %q", msg.GetType().String(), rsp.GetType().String())
	}
	log.Printf("Unmarshalled to type: %q, info: %q", rsp.GetType().String(), rsp.GetInfo())
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
		{
			Name:    "listen",
			Aliases: []string{"l"},
			Usage:   "infinite listen loop",
			Action:  e(listen),
		},
		{
			Name:    "test",
			Aliases: []string{"t"},
			Usage:   "simple test function",
			Action:  test,
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
