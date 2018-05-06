package main

import (
	"bufio"
	"fmt"
	"os"
)

type API struct {
	client *Client
}

const (
	identity = "identity"
	list     = "list"
	relay    = "relay"
	quit     = "quit"
	help     = "help"
)

func NewAPI(client *Client) *API {
	return &API{
		client: client,
	}
}

func (a *API) Run() {
	scanner := bufio.NewScanner(os.Stdin)

	var cmd string
	for cmd != quit {
		fmt.Printf("Enter command: ")
		scanner.Scan()
		cmd = scanner.Text()

		switch cmd {
		case identity:
			id, err := a.client.GetIdentity()
			if err != nil {
				fmt.Printf("GetIdentity failed: %s", err.Error())
				continue
			}
			fmt.Printf("my user_id=%d\n", id)
		case list:
			ids, err := a.client.ListUsers()
			if err != nil {
				fmt.Printf("ListUsers failed: %s", err.Error())
				continue
			}
			fmt.Printf("active users=%v\n", ids)
		case quit:
			break
		case help:
			fmt.Printf(`
Client 

Usage:

identity - authentify on hub (if not already authentified)
quit - quit the program
help - show this help

`)
		default:
			fmt.Println("Unknown command")
		}
	}
}
