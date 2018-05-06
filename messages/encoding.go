package messages

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/gogo/protobuf/proto"
)

type MsgType uint8

const (
	typeLen = 1
	sizeLen = 4

	MsgTypeUnknown MsgType = iota
	MsgTypeRequest
	MsgTypeIdentityResponse
	MsgTypeListResponse
	MsgTypeRelayRequest
	MsgTypeRelay
)

func Encode(msg proto.Marshaler, msgType MsgType) ([]byte, error) {
	b, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("marshalling failed: %s", err.Error())
	}

	length := uint(len(b))
	if length > math.MaxUint32 {
		return nil, fmt.Errorf("message too large (%d bytes)", length)
	}

	var buf = make([]byte, typeLen+sizeLen+len(b))

	buf[0] = byte(msgType)
	// Write length of b into buf
	binary.BigEndian.PutUint32(buf[1:], uint32(length))
	// Copy encoded msg to buf
	copy(buf[5:], b)

	return buf, nil
}

func Decode(r io.Reader) ([]byte, MsgType, error) {
	var msg []byte
	header := [typeLen + sizeLen]byte{}
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return nil, MsgTypeUnknown, err
	}

	length := binary.BigEndian.Uint32(header[1:])

	msgType := MsgType(header[0])

	msg = make([]byte, int(length))
	if _, err := io.ReadFull(r, msg); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, MsgTypeUnknown, err
	}

	return msg, msgType, nil
}
