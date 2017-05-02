// fileportal project main.go
package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

var Input = flag.String("i", "", "Input File")
var Output = flag.String("o", "", "Output File")

func main() {
	flag.Parse()

	if *Input != "" && *Output != "" {
		fmt.Println("You can only specify an input OR an output, not both!")
		os.Exit(0)
	}

	if *Input != "" {
		fileL, _ := net.Listen("tcp", ":0")
		addr := fileL.Addr().String()
		beacon := time.NewTicker(1 * time.Second)
		go func() {
			udpAddr, _ := net.ResolveUDPAddr("udp4", ":6534")
			broadcast, _ := net.ResolveUDPAddr("udp4", "255.255.255.255:6534")
			c, _ := net.ListenUDP("udp4", udpAddr)
			defer c.Close()
			for range beacon.C {
				c.WriteToUDP([]byte(addr), broadcast)
			}
		}()

		c, _ := fileL.Accept()
		fileL.Close()

		beacon.Stop()

		f, _ := os.Open(*Input)

		io.Copy(c, f)

		c.Close()
		f.Close()

	} else if *Output != "" {
		udpAddr, _ := net.ResolveUDPAddr("udp4", ":6534")
		l, _ := net.ListenUDP("udp4", udpAddr)

		var c net.Conn
		for {
			buf := make([]byte, 1500)

			n, _ := l.Read(buf)

			addr := string(buf[:n])
			var err error
			c, err = net.Dial("tcp", addr)
			if err != nil {
				continue
			}
			break
		}

		f, _ := os.Create(*Output)

		io.Copy(f, c)

		f.Close()
	}

}
