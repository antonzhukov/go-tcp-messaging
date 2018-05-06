package messages

import (
	"io"
	"reflect"
	"testing"

	"bytes"

	"github.com/gogo/protobuf/proto"
)

func Test_Encode(t *testing.T) {
	type args struct {
		msg     proto.Marshaler
		msgType MsgType
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			"empty request",
			args{&Request{}, MsgTypeRequest},
			[]byte{1, 0, 0, 0, 0},
			false,
		},
		{
			"request with id",
			args{&Request{Id: 123}, MsgTypeRequest},
			[]byte{1, 0, 0, 0, 2, 16, 123},
			false,
		},
		{
			"identity response",
			args{&IdentityResponse{Id: 123}, MsgTypeIdentityResponse},
			[]byte{2, 0, 0, 0, 2, 8, 123},
			false,
		},
		{
			"list response",
			args{&ListResponse{Ids: []int32{123}}, MsgTypeListResponse},
			[]byte{3, 0, 0, 0, 3, 10, 1, 123},
			false,
		},
		{
			"relay request",
			args{&RelayRequest{Id: 456, Ids: []int32{123}, Body: []byte{99}}, MsgTypeRelayRequest},
			[]byte{4, 0, 0, 0, 9, 8, 200, 3, 18, 1, 123, 26, 1, 99},
			false,
		},
		{
			"relay",
			args{&Relay{Body: []byte{99}}, MsgTypeRelay},
			[]byte{5, 0, 0, 0, 3, 26, 1, 99},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Encode(tt.args.msg, tt.args.msgType)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name     string
		args     args
		want     []byte
		wantType MsgType
		wantErr  bool
	}{
		{
			"empty",
			args{bytes.NewReader([]byte{0, 0, 0, 0, 0})},
			[]byte{},
			MsgTypeUnknown,
			false,
		},
		{
			"request with id",
			args{bytes.NewReader([]byte{1, 0, 0, 0, 2, 16, 123})},
			[]byte{16, 123},
			MsgTypeRequest,
			false,
		},
		{
			"identity response",
			args{bytes.NewReader([]byte{2, 0, 0, 0, 2, 8, 123})},
			[]byte{8, 123},
			MsgTypeIdentityResponse,
			false,
		},
		{
			"list response",
			args{bytes.NewReader([]byte{3, 0, 0, 0, 3, 10, 1, 123})},
			[]byte{10, 1, 123},
			MsgTypeListResponse,
			false,
		},
		{
			"relay request",
			args{bytes.NewReader([]byte{4, 0, 0, 0, 9, 8, 200, 3, 18, 1, 123, 26, 1, 99})},
			[]byte{8, 200, 3, 18, 1, 123, 26, 1, 99},
			MsgTypeRelayRequest,
			false,
		},
		{
			"relay",
			args{bytes.NewReader([]byte{5, 0, 0, 0, 3, 26, 1, 99})},
			[]byte{26, 1, 99},
			MsgTypeRelay,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := Decode(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s. Decode() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s. Decode() got = %v, want %v", tt.name, got, tt.want)
			}
			if got1 != tt.wantType {
				t.Errorf("%s. Decode() got1 = %v, want %v", tt.name, got1, tt.wantType)
			}
		})
	}
}
