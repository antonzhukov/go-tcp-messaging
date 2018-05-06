package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
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
		case relay:
			// collect user ids
			fmt.Printf("Enter comma separated list of users to relay message to: ")
			scanner.Scan()
			usersStr := scanner.Text()
			users := strings.Split(usersStr, ",")
			userIds := make([]int32, 0, len(users))
			for _, user := range users {
				if id, err := strconv.Atoi(user); err == nil {
					userIds = append(userIds, int32(id))
				}
			}

			// read message
			fmt.Printf("Enter message: ")
			scanner.Scan()
			msg := scanner.Text()

			// relay message
			fmt.Printf("sending msg='%s' to users=%s\n", msg, usersStr)
			a.client.RelayRequest(userIds, []byte(msg))
		case quit:
			break
		case help:
			fmt.Printf(`
Client 

Usage:

identity - authentify on hub (if not already authentified)
list - show list of currently active users
relay - relay message to selected users
quit - quit the program
help - show this help

`)
		default:
			fmt.Println("Unknown command")
		}
	}
}
