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
		case messages.MsgTypeRelayRequest:
			h.logger.Info("new relay request")
			h.relayRequest(bytes)
		}
	}
}

func (h *Hub) handleRequest(conn net.Conn, bytes []byte, closeChan chan bool) {
	// parse message
	var request messages.Request
	err := proto.Unmarshal(bytes, &request)
	if err != nil {
		h.logger.Error("unmarshal failed", zap.Error(err))
		return
	}

	switch request.Type {
	case messages.Request_IDENTITY:
		h.logger.Info("new identity request")
		id, err := h.identityRequest(conn)
		if err != nil {
			h.logger.Error("identityRequest failed", zap.Error(err))
			break
		}
		// subscribe all authenticated users to relay events
		go h.subscribeUser(id, conn, closeChan)
	case messages.Request_LIST:
		h.logger.Info("new list request")
		h.listRequest(request.Id, conn)
	}
}

// identityRequest handles request and sends the response with id
func (h *Hub) identityRequest(conn net.Conn) (int32, error) {
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

// listRequest handles request and responds with a list of currently subscribed users
func (h *Hub) listRequest(userID int32, conn net.Conn) {
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

// relayRequest handles relay message and sends it to all currently active users
func (h *Hub) relayRequest(bytes []byte) {
	// parse message
	var request messages.RelayRequest
	err := proto.Unmarshal(bytes, &request)
	if err != nil {
		h.logger.Error("unmarshal failed", zap.Error(err))
		return
	}

	if len(request.Ids) == 0 {
		return
	}
	body := request.Body
	ids := request.Ids
	if len(body) > messages.BodyMaxLength {
		body = body[:messages.BodyMaxLength]
	}

	if len(ids) > messages.MaxReceivers {
		ids = ids[:messages.MaxReceivers]
	}

	// prepare relay message
	relay := &messages.Relay{
		Body: body,
	}
	bytes, err = messages.Encode(relay, messages.MsgTypeRelay)
	if err != nil {
		panic(fmt.Sprintf("Relay marshalling failed, %s", err))
	}

	// send relay
	h.lock.RLock()
	for _, id := range ids {
		if id == request.Id {
			continue
		}

		if conn, ok := h.subscribers[id]; ok {
			if conn != nil {
				go conn.Write(bytes)
			}
		}
	}
	h.lock.RUnlock()
}
