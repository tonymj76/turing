package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

const messageSize = 4

func main() {
	conn, err := net.Dial("tcp", ":1234")
	if err != nil {
		log.Fatalf("Error dialing: %v", err)
	}
	defer conn.Close()

	// Example: Read data from the server
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

	message := 200 // example value to send
	data := make([]byte, messageSize)
	binary.BigEndian.PutUint32(data, uint32(message))

	_, err = conn.Write(data)
	if err != nil {
		log.Printf("Error writing: %v", err)
		return
	}
}
