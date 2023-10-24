package codec

import (
	"bytes"
	"encoding/json"
	"github.com/BabySid/gobase"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
)

type CodecType int

const (
	JsonCodec     CodecType = 0
	ProtobufCodec CodecType = 1
)

type ParamDecoder func(raw json.RawMessage, params interface{}) error
type ReplyEncoder func(reply interface{}) ([]byte, error)

func GetParamDecoder(c CodecType) ParamDecoder {
	switch c {
	case JsonCodec:
		return StdParamsDecoder
	case ProtobufCodec:
		return ProtobufParamsDecoder
	default:
		gobase.AssertHere()
	}
	return nil
}

func GetReplyEncoder(c CodecType) ReplyEncoder {
	switch c {
	case JsonCodec:
		return StdReplyEncoder
	case ProtobufCodec:
		return ProtobufReplyEncoder
	default:
		gobase.AssertHere()
	}
	return nil
}

func StdParamsDecoder(raw json.RawMessage, params interface{}) error {
	err := json.Unmarshal(raw, params)
	return err
}

func StdReplyEncoder(reply interface{}) ([]byte, error) {
	bs, err := json.Marshal(reply)
	if err != nil {
		return nil, err
	}

	return bs, nil
}

var (
	DefaultProtoMarshal = &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			EmitUnpopulated: true,
			UseEnumNumbers:  true,
			UseProtoNames:   true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}
)

func ProtobufParamsDecoder(raw json.RawMessage, params interface{}) error {
	newReader, err := utilities.IOReaderFactory(bytes.NewReader(raw))
	if err != nil {
		return err
	}

	err = DefaultProtoMarshal.NewDecoder(newReader()).Decode(params)
	if err != nil && err != io.EOF {
		return err
	}

	return nil
}

func ProtobufReplyEncoder(reply interface{}) ([]byte, error) {
	return DefaultProtoMarshal.Marshal(reply)
}
