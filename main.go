package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"time"

	"github.com/afking/godelta/delta"
	"github.com/golang/protobuf/proto"

	"github.com/codegangsta/cli"
)

const (
	TIMEOUT time.Duration = time.Second * 4

	CONN_LOCL string = "192.168.1.100"
	CONN_MATL string = "127.0.0.1"

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

/*
func handleConn(conn1 net.Conn) {
	log.Println("Got a request")
	var buf bytes.Buffer
	n, err := io.Copy(&buf, conn)
	if err != nil {
		log.Println("handleConn: ", err)
	}
	log.Printf("%i bytes read\n", n)
}
*/
func read(msg *delta.Message) error {
	var buf bytes.Buffer
	n, err := io.Copy(&buf, conn)
	if err != nil {
		return err
	}
	log.Printf("%d bytes read\n", n)
	return proto.Unmarshal(buf.Bytes(), msg)
	/*
		data := make([]byte, 128)
		//if err := conn.SetReadDeadline(time.Now().Add(TIMEOUT)); err != nil {
		//	return nil, err
		//}
		n, err := conn.Read(data)
		if err != nil {
			return err
		}
		log.Printf("%i bytes readq\n", n)
		return proto.Unmarshal(data, msg)
	*/
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
	log.Printf("%d bytes written\n", n)
	return nil
}

func msgType(t delta.Message_Type) error {
	msg := &delta.Message{
		Type: &t,
	}

	return write(msg)
}

func msgPoint(x, y, z float64) error {
	log.Printf("POINT(%f, %f, %f)", x, y, z)
	msg := &delta.Message{
		Type: delta.Message_POINT.Enum(),
		Point: &delta.Point{
			X: &x,
			Y: &y,
			Z: &z,
		},
	}

	return write(msg)
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

func start(c *cli.Context) error {
	return msgType(delta.Message_START)
}
func stop(c *cli.Context) error {
	return msgType(delta.Message_STOP)
}
func point(c *cli.Context) error {
	return msgPoint(0.02, 0.02, 0.0)
}
func xbox(c *cli.Context) error {
	return xboxDriver()
}
func set(c *cli.Context) error {
	return nil // TODO
}
func get(c *cli.Context) error {
	return msgType(delta.Message_GET)
}
func circle(c *cli.Context) error {
	ts := time.Now()
	for time.Since(ts).Seconds() < math.Pi*4 {
		t := time.Since(ts).Seconds()
		if err := msgPoint(math.Sin(t)*0.04, math.Cos(t)*0.04, 0); err != nil {
			log.Println(err)
		}
		time.Sleep(time.Millisecond * 3)
	}

	if err := msgPoint(0, 0, 0); err != nil {
		return err
	}

	return nil
}
func proxy(c *cli.Context) error {
	TCP() // INIT

	addr, err := net.ResolveUDPAddr("udp", ":8080")
	if err != nil {
		return err
	}
	udp, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	defer udp.Close()

	for {
		var buf bytes.Buffer
		n, err := io.Copy(&buf, udp)
		if err != nil {
			return err
		}
		log.Printf("%d bytes read\n", n)
		fmt.Print(buf.String())

		// MATLAB CONVERSION

		if err := msgPoint(0.01, 0.01, 0.01); err != nil {
			return err
		}
	}

	return nil
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

	log.Printf("Message = %d bytes", len(data))

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
			Name:    "start",
			Aliases: []string{"s"},
			Usage:   "start allows motor positioning commands",
			Action:  e(start),
		},
		{
			Name:    "stop",
			Aliases: []string{"h"},
			Usage:   "stop ignores motor positioning commands",
			Action:  e(stop),
		},
		{
			Name:   "point",
			Usage:  "send point command",
			Action: e(point),
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
		{
			Name:    "set",
			Aliases: []string{"s"},
			Usage:   "set motoro data",
			Action:  e(set),
		},
		{
			Name:    "get",
			Aliases: []string{"g"},
			Usage:   "get motoro data",
			Action:  e(get),
		},
		{
			Name:   "circle",
			Usage:  "make a circle",
			Action: e(circle),
		},
		{
			Name:   "proxy",
			Usage:  "proxy matlab commands to points commands",
			Action: e(proxy),
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
