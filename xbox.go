package main

import (
	"fmt"
	"log"
	//"math"
	"time"

	"github.com/kylelemons/gousb/usb"
	//"github.com/kylelemons/gousb/usbid"
)

type xboxCtrl struct {
	ctx        *usb.Context
	controller *usb.Device
	in         usb.Endpoint
	out        usb.Endpoint

	// Commands
	last [512]byte
	cur  [512]byte

	// JoyStick - 16bit
	xLS float64
	yLS float64
	xRS float64
	yRS float64
}

func (x *xboxCtrl) send() {
	// Format for delta arm
	// 0.04 m radius, 16 bit max = 32768
	dFmt := func(a float64) float64 {
		return a / 32768 * 0.04
	}

	if err := msgPoint(dFmt(x.xLS), dFmt(x.yLS), 0); err != nil {
		log.Println("xbox: ", err)
	}
}

func xboxDriver() error {
	x := &xboxCtrl{}

	// One context should be opened for the application.
	x.ctx = usb.NewContext()
	defer x.ctx.Close()

	// ListDevices is used to find the devices to open.
	devs, err := x.ctx.ListDevices(func(desc *usb.Descriptor) bool {
		if desc.Vendor == usb.ID(0x045e) && desc.Product == usb.ID(0x028e) {
			return true
		} else {
			return false
		}
	})
	if err != nil {
		return err
	} else if len(devs) == 0 {
		return fmt.Errorf("Xbox controller was not detected.")
	}
	defer func() {
		for _, d := range devs {
			d.Close()
		}
	}()

	if len(devs) != 1 {
		return fmt.Errorf("Found %d devices, want 1", len(devs))
	}
	// Setup Device, Grab first controller
	x.controller = devs[0]

	if err := x.controller.Reset(); err != nil {
		return err
	}

	// Open Endpoints
	// config = 1, iface = 0, setup = 0, endIn = 1, endOut = 1

	x.in, err = x.controller.OpenEndpoint(01, 00, 00, 01|uint8(usb.ENDPOINT_DIR_IN))
	if err != nil {
		return err
	}

	x.out, err = x.controller.OpenEndpoint(01, 00, 00, 01|uint8(usb.ENDPOINT_DIR_OUT))
	if err != nil {
		return err
	}

	x.xbox360()
	return nil
}

const (
	Empty      byte = iota // 00000000 ( 0) no LEDs
	WarnAll                // 00000001 ( 1) flash all briefly
	NewPlayer1             // 00000010 ( 2) p1 flash then solid
	NewPlayer2             // 00000011
	NewPlayer3             // 00000100
	NewPlayer4             // 00000101
	Player1                // 00000110 ( 6) p1 solid
	Player2                // 00000111
	Player3                // 00001000
	Player4                // 00001001
	Waiting                // 00001010 (10) empty w/ loops
	WarnPlayer             // 00001011 (11) flash active
	_                      // 00001100 (12) empty
	Battery                // 00001101 (13) squiggle
	Searching              // 00001110 (14) slow flash
	Booting                // 00001111 (15) solid then flash
)

func (x *xboxCtrl) led(b byte) {
	x.out.Write([]byte{0x01, 0x03, b})
}

func (x *xboxCtrl) setPlayer(player byte) {
	spin := []byte{
		Player1, Player2, Player4, Player3,
	}
	spinIdx := 0
	spinDelay := 100 * time.Millisecond

	x.led(Booting)
	time.Sleep(100 * time.Millisecond)
	for spinDelay > 20*time.Millisecond {
		x.led(spin[spinIdx])
		time.Sleep(spinDelay)
		spinIdx = (spinIdx + 1) % len(spin)
		spinDelay -= 5 * time.Millisecond
	}
	for i := 0; i < 40; i++ { // just for safety
		cur := spin[spinIdx]
		x.led(cur)
		time.Sleep(spinDelay)
		spinIdx = (spinIdx + 1) % len(spin)
		if cur == player {
			break
		}
	}
}

func (x *xboxCtrl) dword(hi, lo byte) int16 {
	return int16(hi)<<8 | int16(lo)
}

func (x *xboxCtrl) decode() {
	n, err := x.in.Read(x.cur[:])
	if err != nil || n != 20 {
		log.Printf("ignoring read: %d bytes, err = %v", n, err)
		return
	}

	// 1-bit values
	for _, v := range []struct {
		idx  int
		bit  uint
		name string
	}{
		{2, 0, "DPAD U"},
		{2, 1, "DPAD D"},
		{2, 2, "DPAD L"},
		{2, 3, "DPAD R"},
		{2, 4, "START"},
		{2, 5, "BACK"},
		{2, 6, "THUMB L"},
		{2, 7, "THUMB R"},
		{3, 0, "LB"},
		{3, 1, "RB"},
		{3, 2, "GUIDE"},
		{3, 4, "A"},
		{3, 5, "B"},
		{3, 6, "X"},
		{3, 7, "Y"},
	} {
		c := x.cur[v.idx] & (1 << v.bit)
		l := x.last[v.idx] & (1 << v.bit)
		if c == l {
			continue
		}
		switch {
		case c != 0:
			log.Printf("Button %q pressed", v.name)
		case l != 0:
			log.Printf("Button %q released", v.name)
		}
	}

	// 8-bit values
	for _, v := range []struct {
		idx  int
		name string
	}{
		{4, "LT"},
		{5, "RT"},
	} {
		c := x.cur[v.idx]
		l := x.last[v.idx]
		if c == l {
			continue
		}
		log.Printf("Trigger %q = %v", v.name, c)
	}

	/*
		//     +y
		//      N
		// -x W-|-E +x
		//      S
		//     -y
		dirs := [...]string{
			"W", "SW", "S", "SE", "E", "NE", "N", "NW", "W",
		}
		dir := func(x, y int16) (string, int32) {
			// Direction
			rad := math.Atan2(float64(y), float64(x))
			dir := 4 * rad / math.Pi
			card := int(dir + math.Copysign(0.5, dir))

			// Magnitude
			mag := math.Sqrt(float64(x)*float64(x) + float64(y)*float64(y))
			return dirs[card+4], int32(mag)
		} */

	// 16-bit values
	// LS
	x.xLS = float64(x.dword(x.cur[7], x.cur[6]))
	x.yLS = float64(x.dword(x.cur[9], x.cur[8]))

	// RS
	x.xRS = float64(x.dword(x.cur[11], x.cur[10]))
	x.yRS = float64(x.dword(x.cur[13], x.cur[12]))

	/*
		for _, v := range []struct {
			hiX, loX int
			hiY, loY int
			name     string
		}{
			{7, 6, 9, 8, "LS"},
			{11, 10, 13, 12, "RS"},
		} {
			c, cmag := dir(
				x.dword(x.cur[v.hiX], x.cur[v.loX]),
				x.dword(x.cur[v.hiY], x.cur[v.loY]),
			)
			l, lmag := dir(
				x.dword(x.last[v.hiX], x.last[v.loX]),
				x.dword(x.last[v.hiY], x.last[v.loY]),
			)
			ccenter := cmag < 10240
			lcenter := lmag < 10240
			if ccenter && lcenter {
				continue
			}
			if c == l && cmag == lmag {
				continue
			}
			if cmag > 10240 {
				log.Printf("Stick %q = %v x %v", v.name, c, cmag)
			} else {
				log.Printf("Stick %q centered", v.name)
			}
		}
	*/
	x.last, x.cur = x.cur, x.last
}

func (x *xboxCtrl) xbox360() {
	// https://github.com/Grumbel/xboxdrv/blob/master/PROTOCOL
	x.led(Empty)
	time.Sleep(1 * time.Second)
	x.setPlayer(Player1)

	var b [512]byte
	for {
		n, err := x.in.Read(b[:])
		log.Printf("read %d bytes: % x [err: %v]", n, b[:n], err)
		if err != nil {
			break
		}
	}

	x.controller.ReadTimeout = 60 * time.Second
	for {
		x.decode()
		x.send()
		time.Sleep(time.Millisecond * 10)
	}
}
