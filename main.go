// fileportal project main.go
package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"time"
)

var Input = flag.String("i", "", "Input File")

type FileData struct {
	Name string
	Size int64
	Port int
}

func main() {
	flag.Parse()

	if *Input != "" {
		f, err := os.Open(*Input)
		if err != nil {
			log.Fatalf("Error opening input file: %s", err)
		}

		fileL, err := net.Listen("tcp", ":0")
		if err != nil {
			log.Fatalf("Error opening tcp server: %s", err)
		}

		port := fileL.Addr().(*net.TCPAddr).Port

		info, err := f.Stat()
		if err != nil {
			log.Fatalf("Error stat'ing input file: %s", err)
		}

		fileData := FileData{}
		fileData.Name = info.Name()
		fileData.Size = info.Size()
		fileData.Port = port

		filePacket, _ := json.Marshal(fileData)

		beacon := time.NewTicker(1 * time.Second)
		go func() {
			udpAddr, _ := net.ResolveUDPAddr("udp4", ":0")
			broadcast, _ := net.ResolveUDPAddr("udp4", "255.255.255.255:6534")
			c, _ := net.ListenUDP("udp4", udpAddr)
			defer c.Close()
			for range beacon.C {
				c.WriteToUDP(filePacket, broadcast)
			}
		}()

		c, _ := fileL.Accept()
		fileL.Close()

		log.Printf("Client connected: %s", c.RemoteAddr())

		beacon.Stop()

		io.Copy(c, f)

		c.Close()
		f.Close()

	} else {
		udpAddr, _ := net.ResolveUDPAddr("udp4", ":6534")
		l, _ := net.ListenUDP("udp4", udpAddr)

		var c net.Conn
		fileData := FileData{}

		for {
			buf := make([]byte, 1500)

			n, addr, _ := l.ReadFromUDP(buf)

			err := json.Unmarshal(buf[:n], &fileData)
			if err != nil {
				continue
			}

			c, err = net.Dial("tcp", (&net.TCPAddr{IP: addr.IP, Port: fileData.Port}).String())
			if err == nil {
				log.Printf("Connected to client, downloading file: %s", fileData.Name)
				break
			}
		}

		f, _ := os.Create(fileData.Name)

		io.Copy(f, c)

		f.Close()
	}

}
