package main

import (
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
)

const (
	port           = 8888
	requestTimeout = 5 * time.Second
)

func main() {
	// init logger
	l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	// init hub connection
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		panic(fmt.Sprintf("Dial failed: %s", err.Error()))
	}

	// init client
	client := NewClient(l, conn, requestTimeout)
	err = client.Run()
	if err != nil {
		conn.Close()
		panic(fmt.Sprintf("client.Run failed: %s", err.Error()))
	}

	// init command line interface
	api := NewAPI(client)
	api.Run()
}
