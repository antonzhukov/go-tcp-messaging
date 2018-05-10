package main

import (
	"net"
	"reflect"
	"testing"
	"github.com/antonzhukov/go-tcp-messaging/messages"

	"github.com/gogo/protobuf/proto"
)

func TestClient_receiveMessages(t *testing.T) {
	// arrange
	server, client := net.Pipe()
	responseChan := make(chan messageRaw, 1)
	c := &Client{
		conn:         client,
		responseChan: responseChan,
	}
	go c.receiveMessages()

	// act
	idResp := &messages.IdentityResponse{
		Id: 123,
	}
	bytes, err := messages.Encode(idResp, messages.MsgTypeIdentityResponse)
	if err != nil {
		t.Error(err)
	}
	server.Write(bytes)

	// assert
	msgRaw := <-c.responseChan
	if msgRaw.msgType != messages.MsgTypeIdentityResponse {
		t.Errorf("receiveMessages failed. Expected %d, got %d", messages.MsgTypeIdentityResponse, msgRaw.msgType)
	}

	result := &messages.IdentityResponse{}
	err = proto.Unmarshal(msgRaw.msg, result)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result, idResp) {
		t.Errorf("receiveMessages failed. Expected %#v, got %#v", idResp, result)
	}
}

func TestClient_GetIdentity(t *testing.T) {
	// arrange
	server, client := net.Pipe()
	c := &Client{
		conn:         client,
		responseChan: make(chan messageRaw, 1),
	}
	resultChan := make(chan int32)

	// act
	go func(resultChan chan int32) {
		res, err := c.GetIdentity()
		if err != nil {
			t.Error(err)
		}
		resultChan <- res
	}(resultChan)

	// assert request
	bytes, msgType, err := messages.Decode(server)
	if err != nil {
		t.Errorf("GetIdentity failed. Unexpected err: %s", err.Error())
	}
	var result messages.Request
	err = proto.Unmarshal(bytes, &result)
	if err != nil {
		t.Error(err)
	}

	if msgType != messages.MsgTypeRequest {
		t.Errorf("GetIdentity failed. Expected %d, got %d", messages.MsgTypeRequest, msgType)
	}

	expectedReq := messages.Request{
		Type: messages.Request_IDENTITY,
	}
	if !reflect.DeepEqual(result, expectedReq) {
		t.Errorf("GetIdentity failed. Expected %#v, got %#v", expectedReq, result)
	}

	// assert user id
	response := &messages.IdentityResponse{
		Id: 123,
	}
	bytes, err = response.Marshal()
	if err != nil {
		t.Error(err)
	}
	c.responseChan <- messageRaw{bytes, messages.MsgTypeIdentityResponse}
	id := <-resultChan

	if id != 123 {
		t.Errorf("GetIdentity failed. Expected %d, got %d", 123, id)
	}
}

func TestClient_ListUsers(t *testing.T) {
	// arrange
	server, client := net.Pipe()
	c := &Client{
		conn:         client,
		responseChan: make(chan messageRaw, 1),
	}
	resultChan := make(chan []int32)

	// act
	go func(resultChan chan []int32) {
		res, err := c.ListUsers()
		if err != nil {
			t.Error(err)
		}
		resultChan <- res
	}(resultChan)

	// assert request
	bytes, msgType, err := messages.Decode(server)
	if err != nil {
		t.Errorf("ListUsers failed. Unexpected err: %s", err.Error())
	}
	var result messages.Request
	err = proto.Unmarshal(bytes, &result)
	if err != nil {
		t.Error(err)
	}

	if msgType != messages.MsgTypeRequest {
		t.Errorf("ListUsers failed. Expected %d, got %d", messages.MsgTypeRequest, msgType)
	}

	expectedReq := messages.Request{
		Type: messages.Request_LIST,
	}
	if !reflect.DeepEqual(result, expectedReq) {
		t.Errorf("ListUsers failed. Expected %#v, got %#v", expectedReq, result)
	}

	// assert user ids
	response := &messages.ListResponse{
		Ids: []int32{123, 456},
	}
	bytes, err = response.Marshal()
	if err != nil {
		t.Error(err)
	}
	c.responseChan <- messageRaw{bytes, messages.MsgTypeListResponse}
	ids := <-resultChan

	if !reflect.DeepEqual(response.Ids, ids) {
		t.Errorf("GetIdentity failed. Expected %#v, got %#v", response.Ids, ids)
	}
}

func TestClient_RelayRequest(t *testing.T) {
	// arrange
	server, client := net.Pipe()
	c := &Client{
		conn: client,
	}

	// act
	ids := []int32{123, 456}
	msg := "Hello go"
	go c.RelayRequest(ids, []byte(msg))

	// assert
	bytes, msgType, err := messages.Decode(server)
	if err != nil {
		t.Errorf("RelayRequest failed. Unexpected err: %s", err.Error())
	}
	var result messages.RelayRequest
	err = proto.Unmarshal(bytes, &result)
	if err != nil {
		t.Error(err)
	}

	if msgType != messages.MsgTypeRelayRequest {
		t.Errorf("RelayRequest failed. Expected %d, got %d", messages.MsgTypeRelayRequest, msgType)
	}

	expectedReq := messages.RelayRequest{
		Ids:  ids,
		Body: []byte(msg),
	}
	if !reflect.DeepEqual(result, expectedReq) {
		t.Errorf("RelayRequest failed. Expected %#v, got %#v", expectedReq, result)
	}
}
