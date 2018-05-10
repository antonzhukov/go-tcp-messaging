package main

import (
	"testing"

	"net"
	"github.com/antonzhukov/go-tcp-messaging/messages"

	"reflect"

	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap"
)

func TestHub_identityRequest(t *testing.T) {
	// arrange
	server, client := net.Pipe()
	h := &Hub{
		subscribers:   make(map[int32]net.Conn),
		ln:            &mockListener{conn: server},
		logger:        zap.L(),
		usersProvider: NewUsers(),
	}
	go h.handleConnection(server)

	// act
	relayReq := &messages.Request{
		Type: messages.Request_IDENTITY,
	}
	bytes, err := messages.Encode(relayReq, messages.MsgTypeRequest)
	if err != nil {
		t.Error(err)
	}
	go client.Write(bytes)

	// assert
	bytes, msgType, err := messages.Decode(client)
	if msgType != messages.MsgTypeIdentityResponse {
		t.Errorf("identityRequest failed. Expected %d, got %d", messages.MsgTypeIdentityResponse, msgType)
	}

	expectedResp := messages.IdentityResponse{
		Id: 1,
	}
	var result messages.IdentityResponse
	err = proto.Unmarshal(bytes, &result)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedResp, result) {
		t.Errorf("identityRequest failed. Expected %#v, got %#v", expectedResp, result)
	}
}

func TestHub_listRequest(t *testing.T) {
	// arrange
	server, client := net.Pipe()
	h := &Hub{
		subscribers:   make(map[int32]net.Conn),
		ln:            &mockListener{conn: server},
		logger:        zap.L(),
		usersProvider: NewUsers(),
	}
	h.subscribers[234] = server
	h.subscribers[435] = server
	go h.handleConnection(server)

	// act
	listReq := &messages.Request{
		Type: messages.Request_LIST,
	}
	bytes, err := messages.Encode(listReq, messages.MsgTypeRequest)
	if err != nil {
		t.Error(err)
	}
	go client.Write(bytes)

	// assert
	bytes, msgType, err := messages.Decode(client)
	if msgType != messages.MsgTypeListResponse {
		t.Errorf("listRequest failed. Expected %d, got %d", messages.MsgTypeListResponse, msgType)
	}

	expectedResp := messages.ListResponse{
		Ids: []int32{234, 435},
	}
	var result messages.ListResponse
	err = proto.Unmarshal(bytes, &result)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedResp, result) {
		t.Errorf("listRequest failed. Expected %#v, got %#v", expectedResp, result)
	}
}

func TestHub_relayRequest(t *testing.T) {
	// arrange
	server, client := net.Pipe()
	h := &Hub{
		subscribers: make(map[int32]net.Conn),
		ln:          &mockListener{conn: server},
		logger:      zap.L(),
	}
	h.subscribers[123] = server
	go h.handleConnection(server)

	// act
	relayReq := &messages.RelayRequest{
		Id:   456,
		Ids:  []int32{123},
		Body: []byte("g'day"),
	}
	bytes, err := messages.Encode(relayReq, messages.MsgTypeRelayRequest)
	if err != nil {
		t.Error(err)
	}
	go client.Write(bytes)

	// assert
	bytes, msgType, err := messages.Decode(client)
	if msgType != messages.MsgTypeRelay {
		t.Errorf("relayRequest failed. Expected %d, got %d", messages.MsgTypeRelay, msgType)
	}

	expectedRelay := messages.Relay{
		Body: relayReq.Body,
	}
	var result messages.Relay
	err = proto.Unmarshal(bytes, &result)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedRelay, result) {
		t.Errorf("relayRequest failed. Expected %#v, got %#v", expectedRelay, result)
	}

}

type mockListener struct {
	conn net.Conn
}

func (m *mockListener) Accept() (net.Conn, error) {
	return m.conn, nil
}

func (m *mockListener) Close() error {
	return nil
}

func (m *mockListener) Addr() net.Addr {
	return nil
}
