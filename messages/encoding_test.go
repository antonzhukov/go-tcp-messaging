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
			[]byte{3, 0, 0, 0, 0},
			false,
		},
		{
			"request with id",
			args{&Request{Id: 123}, MsgTypeRequest},
			[]byte{3, 0, 0, 0, 2, 16, 123},
			false,
		},
		{
			"identity response",
			args{&IdentityResponse{Id: 123}, MsgTypeIdentityResponse},
			[]byte{4, 0, 0, 0, 2, 8, 123},
			false,
		},
		{
			"list response",
			args{&ListResponse{Ids: []int32{123}}, MsgTypeListResponse},
			[]byte{5, 0, 0, 0, 3, 10, 1, 123},
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
			nil,
			MsgTypeUnknown,
			false,
		},
		{
			"request with id",
			args{bytes.NewReader([]byte{3, 0, 0, 0, 2, 16, 123})},
			[]byte{16, 123},
			MsgTypeRequest,
			false,
		},
		{
			"identity response",
			args{bytes.NewReader([]byte{4, 0, 0, 0, 2, 8, 123})},
			[]byte{8, 123},
			MsgTypeIdentityResponse,
			false,
		},
		{
			"list response",
			args{bytes.NewReader([]byte{5, 0, 0, 0, 3, 10, 1, 123})},
			[]byte{10, 1, 123},
			MsgTypeListResponse,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := Decode(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Decode() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.wantType {
				t.Errorf("Decode() got1 = %v, want %v", got1, tt.wantType)
			}
		})
	}
}
