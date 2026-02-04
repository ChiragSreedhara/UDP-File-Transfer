//run this then receiver (go run sender.go && go run receiver.go, in seperate terminals)
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
	"hash/crc32"
)

type Header struct {
	SequenceID   uint32
	TotalPackets uint32
	DataSize     uint16
	Checksum     uint32
}

func (h *Header) Serialize() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, h)
	return buf.Bytes()
}

func startAckListener(conn *net.UDPConn, ackChan chan<- uint32) {
	for {
		buf := make([]byte, 8)
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Listener error:", err)
			return
		}
		var receivedID uint32
		binary.Read(bytes.NewReader(buf[:n]), binary.BigEndian, &receivedID)
		ackChan <- receivedID
	}
}

func main() {
	serverAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:9999")
	conn, _ := net.DialUDP("udp", nil, serverAddr)
	defer conn.Close()

	ackChan := make(chan uint32)
	go startAckListener(conn, ackChan)

	for i := uint32(1); i <= 5; i++ {
		data := []byte("I love the docker whale :D")
		h := Header{
			SequenceID:   i,
			TotalPackets: 5,
			DataSize:     uint16(len(data)),
			Checksum:     crc32.ChecksumIEEE(data),
		}

		headerBytes := h.Serialize()
		payload := append(headerBytes, data...)		
		fmt.Printf("Sending packet %d\n", i)
		sendPacket(conn, payload, i, ackChan)
	}
}

func sendPacket(conn *net.UDPConn, packet []byte, id uint32, ackChan <-chan uint32) {
	for {
		conn.Write(packet)
		select {
		case receivedID := <-ackChan:
			if receivedID == id {
				fmt.Printf("Packet %d ACK\n", id)
				return
			}
		case <-time.After(2 * time.Second):
			fmt.Printf("timed out %d, resending...\n", id)
		}
	}
}