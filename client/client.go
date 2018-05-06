package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"unity_hw/messages"

	"errors"
	"time"

	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap"
)

type Client struct {
	conn           net.Conn
	responseChan   chan messageRaw
	requestTimeout time.Duration
	logger         *zap.Logger

	id         int32
	identified bool
}

type messageRaw struct {
	msg     []byte
	msgType messages.MsgType
}

func NewClient(logger *zap.Logger, conn net.Conn, requestTimeout time.Duration) *Client {
	return &Client{
		conn:           conn,
		responseChan:   make(chan messageRaw, 1),
		requestTimeout: requestTimeout,
		logger:         logger,
	}
}

func (c *Client) Run() error {
	// start asynchronous receiving messages
	go c.receiveMessages()

	err := c.authenticate()
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) authenticate() error {
	// Get my user ID
	id, err := c.GetIdentity()
	if err != nil {
		return fmt.Errorf("GetIdentity failed: %s", err.Error())
	}
	c.id = id
	c.identified = true
	c.logger.Info("authentication passed", zap.Int32("id", c.id))

	return nil
}

// GetIdentity passes authentication on hub
func (c *Client) GetIdentity() (int32, error) {
	// only authenticate once
	if c.identified {
		return c.id, nil
	}

	// send request
	idReq := &messages.Request{
		Type: messages.Request_IDENTITY,
	}
	bytes, err := messages.Encode(idReq, messages.MsgTypeRequest)
	if err != nil {
		return 0, fmt.Errorf("request marshalling failed, %s", err.Error())
	}
	c.conn.Write(bytes)

	// receive response
	var msgRaw messageRaw
	select {
	case <-time.After(c.requestTimeout):
		return 0, errors.New("identity request timed out")
	case msgRaw = <-c.responseChan:
	}

	if msgRaw.msgType != messages.MsgTypeIdentityResponse {
		return 0, fmt.Errorf("bad response, expected: %d, got %d", messages.MsgTypeIdentityResponse, msgRaw.msgType)
	}
	var idResp messages.IdentityResponse
	proto.Unmarshal(msgRaw.msg, &idResp)

	return idResp.Id, nil
}

// ListUsers returns list of currently active users
func (c *Client) ListUsers() ([]int32, error) {
	// send request
	listReq := &messages.Request{
		Id:   c.id,
		Type: messages.Request_LIST,
	}
	bytes, err := messages.Encode(listReq, messages.MsgTypeRequest)
	if err != nil {
		return nil, fmt.Errorf("request marshalling failed, %s", err.Error())
	}
	c.conn.Write(bytes)

	// receive response
	var msgRaw messageRaw
	select {
	case <-time.After(c.requestTimeout):
		return nil, errors.New("list request timed out")
	case msgRaw = <-c.responseChan:
	}

	if msgRaw.msgType != messages.MsgTypeListResponse {
		return nil, fmt.Errorf("bad response, expected: %d, got %d", messages.MsgTypeListResponse, msgRaw.msgType)
	}
	var listResp messages.ListResponse
	proto.Unmarshal(msgRaw.msg, &listResp)

	return listResp.Ids, nil
}

// RelayRequest relays a message to other users
func (c *Client) RelayRequest(ids []int32, body []byte) error {
	if len(body) > messages.BodyMaxLength {
		body = body[:messages.BodyMaxLength]
	}

	if len(ids) > messages.MaxReceivers {
		ids = ids[:messages.MaxReceivers]
	}
	relayReq := &messages.RelayRequest{
		Id:   c.id,
		Ids:  ids,
		Body: body,
	}
	bytes, err := messages.Encode(relayReq, messages.MsgTypeRelayRequest)
	if err != nil {
		return fmt.Errorf("request marshalling failed, %s", err.Error())
	}
	c.conn.Write(bytes)

	return nil
}

func (c *Client) receiveMessages() {
	bufReader := bufio.NewReader(c.conn)

	for {
		bytes, msgType, err := messages.Decode(bufReader)
		if err == io.EOF {
			panic("connection lost")
		}
		if err != nil {
			c.logger.Error("receiving message failed", zap.Error(err))
			return
		}

		switch msgType {
		case messages.MsgTypeRelay:
			c.handleRelay(bytes)
		case messages.MsgTypeUnknown:
			c.logger.Info("received unknown message, skipping")
		default:
			c.responseChan <- messageRaw{msg: bytes, msgType: msgType}
		}
	}
}

func (c *Client) handleRelay(bytes []byte) {
	// decode relay
	var relay messages.Relay
	proto.Unmarshal(bytes, &relay)
	c.logger.Info("relay message", zap.ByteString("body", relay.Body))
}
