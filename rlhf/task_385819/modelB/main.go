package main

import (
	"encoding/binary"
	"log"
	"net"
)

// PlayerPosition represents the position of a player in the game
type PlayerPosition struct {
	X int32
	Y int32
}

// Server code
func main() {
	// Listen for incoming connections
	ln, err := net.Listen("tcp", ":12345")
	if err != nil {
		log.Fatalf("Error listening: %v", err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		defer conn.Close()

		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	// Read player position data from the client
	pos := new(PlayerPosition)
	err := binary.Read(conn, binary.BigEndian, pos)
	if err != nil {
		log.Printf("Error reading position from client: %v", err)
		return
	}

	// Log the received position
	log.Printf("Received position: %+v", pos)
}
