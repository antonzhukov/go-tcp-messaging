package main

import (
	"net"

	"fmt"

	"go.uber.org/zap"
)

const (
	port = 8888
)

func main() {
	// init logger
	l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	// initialize listener
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	defer func() {
		ln.Close()
	}()

	l.Info("Listening for requests", zap.Int("port", port))

	// initialize hub
	hub := NewHub(l, ln)
	hub.Run()
}
