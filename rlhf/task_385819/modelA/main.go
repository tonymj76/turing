package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

const messageSize = 4

func main() {
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatalf("Error listening: %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	message := 100 // example value to send
	data := make([]byte, messageSize)
	binary.BigEndian.PutUint32(data, uint32(message))

	_, err := conn.Write(data)
	if err != nil {
		log.Printf("Error writing: %v", err)
		return
	}

	// Example: Read data back from the client
	buff := make([]byte, messageSize)
	n, err := conn.Read(buff)
	if err != nil {
		log.Printf("Error reading: %v", err)
		return
	}

	if n != messageSize {
		log.Printf("Invalid read size: %d", n)
		return
	}

	receivedMessage := binary.BigEndian.Uint32(buff)
	fmt.Printf("Received message: %d\n", receivedMessage)
}
