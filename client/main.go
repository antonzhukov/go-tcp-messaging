package main

import (
	"bufio"
	"io"
	"log"
	"net"
)

func main() {
	// Init hub connection
	conn, err := net.Dial("tcp", "localhost:8888")
	if err != nil {
		log.Println(err)
		return
	}

	conn.Write([]byte("guten morgen\n"))

	handleIncomingMessages(conn)
}

func handleIncomingMessages(conn net.Conn) {
	bufReader := bufio.NewReader(conn)

	for {
		// Read tokens delimited by newline
		bytes, err := bufReader.ReadBytes('\n')
		if err == io.EOF {
			log.Println("connection lost")
			break
		}
		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("%s", bytes)
	}
}
