package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

const port = ":8888"

func main() {
	// Initialize listener
	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Printf("Listen failed, %s", err.Error())
		return
	}
	defer func() {
		ln.Close()
	}()
	log.Printf("Listening on port %s\n", port)

	// Accept connections
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			break
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	log.Println("new client connected...")

	// Close connection when this function ends
	defer func() {
		log.Println("client disconnected...")
		conn.Close()
	}()

	bufReader := bufio.NewReader(conn)

	for {
		// Read tokens delimited by newline
		bytes, err := bufReader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("%s", bytes)
	}
}
