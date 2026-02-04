package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"math/rand"
	"net"
)

func main() {
	addr, _ := net.ResolveUDPAddr("udp", ":9999")
	conn, _ := net.ListenUDP("udp", addr)
	defer conn.Close()
	fmt.Println("listening on :9999...\n")

	for {
		buf := make([]byte, 1024)
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("read err-", err)
			continue
		}

		if rand.Float32() < 0.33 {
			fmt.Println("packet dropped :(\n")
			continue
		}
		if rand.Float32() < 0.33 {
			fmt.Println("packet corrupted D:\n")
			if n > 14 {
				buf[14] = buf[14] + 1 
			}
		}

		if n < 14 {
			continue
		}

		sentChecksum := binary.BigEndian.Uint32(buf[10:14])
		dataSize := binary.BigEndian.Uint16(buf[8:10])
		receivedData := buf[14 : 14+int(dataSize)]
		calculatedChecksum := crc32.ChecksumIEEE(receivedData)

		if sentChecksum != calculatedChecksum {
			fmt.Println("corruption found")
			continue 
		}

		seqID := binary.BigEndian.Uint32(buf[:4])
		message := string(receivedData) 

		fmt.Printf("Received Packet %d %s \n", seqID, message)

		ack := make([]byte, 4)
		binary.BigEndian.PutUint32(ack, seqID)
		conn.WriteToUDP(ack, remoteAddr)
	}
}