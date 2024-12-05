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

// Client code
func main() {
	// Create a player position to send to the server
	pos := PlayerPosition{
		X: 100,
		Y: 200,
	}

	// Connect to the server
	conn, err := net.Dial("tcp", "127.0.0.1:12345")
	if err != nil {
		log.Fatalf("Error connecting to server: %v", err)
	}
	defer conn.Close()

	// Write the player position data to the server in big-endian format
	err = binary.Write(conn, binary.BigEndian, &pos)
	if err != nil {
		log.Printf("Error writing position to server: %v", err)
		return
	}

	log.Printf("Sent position: %+v", pos)
}
