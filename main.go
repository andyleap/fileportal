// fileportal project main.go
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
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
		port := fileL.Addr().(*net.TCPAddr).Port
		addrs, _ := net.InterfaceAddrs()
		beacon := time.NewTicker(1 * time.Second)
		go func() {

			udpAddr, _ := net.ResolveUDPAddr("udp4", ":6534")
			broadcast, _ := net.ResolveUDPAddr("udp4", "255.255.255.255:6534")
			c, _ := net.ListenUDP("udp4", udpAddr)
			defer c.Close()
			for range beacon.C {
				if len(addrs) == 0 {
					addrs, _ = net.InterfaceAddrs()
				}

				addr := (&net.TCPAddr{IP: addrs[0].(*net.IPNet).IP, Port: port}).String()
				c.WriteToUDP([]byte(addr), broadcast)
				addrs = addrs[1:]
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

			n, _, _ := l.ReadFromUDP(buf)

			log.Println("Message: ", string(buf[:n]))

			addr := string(buf[:n])
			var err error
			c, err = net.Dial("tcp", addr)
			if err != nil {
				log.Println("Dial error: ", err)
				continue
			}
			break
		}

		f, _ := os.Create(*Output)

		io.Copy(f, c)

		f.Close()
	}

}
