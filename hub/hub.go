package main

import (
	"bufio"
	"fmt"
	"net"
	"unity_hw/messages"

	"io"

	"sync"

	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap"
)

type Hub struct {
	ln            net.Listener
	usersProvider UserProvider
	subscribers   map[int32]net.Conn
	lock          sync.RWMutex
	logger        *zap.Logger
}

func NewHub(logger *zap.Logger, ln net.Listener) *Hub {
	return &Hub{
		ln:            ln,
		usersProvider: NewUsers(),
		subscribers:   make(map[int32]net.Conn),
		logger:        logger,
	}
}

func (h *Hub) Run() {
	// Accept connections
	for {
		conn, err := h.ln.Accept()
		if err != nil {
			fmt.Println(err)
			break
		}

		go h.handleConnection(conn)
	}
}

// handleConnection reads requests from users
func (h *Hub) handleConnection(conn net.Conn) {
	// Close connection when this function ends
	defer func() {
		conn.Close()
	}()

	bufReader := bufio.NewReader(conn)
	// use channel to notify of closed connection
	closeChan := make(chan bool, 1)

	for {
		bytes, msgType, err := messages.Decode(bufReader)
		if err == io.EOF {
			closeChan <- true
			break
		}
		if err != nil {
			h.logger.Error("receiving message failed", zap.Error(err))
			return
		}

		switch msgType {
		case messages.MsgTypeUnknown:
			h.logger.Info("received unknown message, skipping")
		case messages.MsgTypeRequest:
			h.handleRequest(conn, bytes, closeChan)
		}
	}
}

func (h *Hub) handleRequest(conn net.Conn, bytes []byte, closeChan chan bool) {
	var request messages.Request
	proto.Unmarshal(bytes, &request)

	switch request.Type {
	case messages.Request_IDENTITY:
		h.logger.Info("new identity request")
		id, err := h.identityCommand(conn)
		if err != nil {
			h.logger.Error("identityCommand failed", zap.Error(err))
			break
		}
		// subscribe all authenticated users to relay events
		go h.subscribeUser(id, conn, closeChan)
	case messages.Request_LIST:
		h.logger.Info("new list request")
		h.listCommand(request.Id, conn)
	}
}

func (h *Hub) identityCommand(conn net.Conn) (int32, error) {
	// authenticate user and handle connection
	id := h.usersProvider.AuthenticateNewUser()
	idResp := &messages.IdentityResponse{
		Id: id,
	}

	bytes, err := messages.Encode(idResp, messages.MsgTypeIdentityResponse)
	if err != nil {
		return 0, fmt.Errorf("encode failed, %s", err.Error())
	}
	conn.Write(bytes)

	return id, nil
}

// subscribeUser subscribes user to relay messages
func (h *Hub) subscribeUser(userID int32, conn net.Conn, closeChan chan bool) {
	h.lock.Lock()
	h.subscribers[userID] = conn
	h.logger.Info("subscribers", zap.Any("list", h.subscribers))
	h.lock.Unlock()

	// unsubscribe user from relay messages if connection is lost
	<-closeChan
	h.lock.Lock()
	delete(h.subscribers, userID)
	h.logger.Info("subscribers", zap.Any("list", h.subscribers))
	h.lock.Unlock()
}

// listCommand responds with a list of currently subscribed users
func (h *Hub) listCommand(userID int32, conn net.Conn) {
	h.lock.RLock()
	ids := make([]int32, 0, len(h.subscribers))
	for id := range h.subscribers {
		if id != userID {
			ids = append(ids, id)
		}
	}
	h.lock.RUnlock()

	listResp := &messages.ListResponse{
		Ids: ids,
	}

	bytes, err := messages.Encode(listResp, messages.MsgTypeListResponse)
	if err != nil {
		panic(fmt.Sprintf("ListResponse marshalling failed, %s", err))
	}
	conn.Write(bytes)

}
