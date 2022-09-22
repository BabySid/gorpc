package httpcfg

import (
	"github.com/BabySid/gobase"
	"github.com/BabySid/gorpc/http/codec"
)

type CodecType int

const (
	JsonCodec     CodecType = 0
	ProtobufCodec CodecType = 1
)

type ServerOption struct {
	Codec CodecType
}

var (
	DefaultOption = ServerOption{Codec: JsonCodec}
)

type ParamDecoder func(raw interface{}, params interface{}) error
type ReplyEncoder func(reply interface{}) ([]byte, error)

func GetParamDecoder(c CodecType) ParamDecoder {
	switch c {
	case JsonCodec:
		return codec.StdParamsDecoder
	case ProtobufCodec:
		return codec.ProtobufParamsDecoder
	default:
		gobase.AssertHere()
	}
	return nil
}

func GetReplyEncoder(c CodecType) ReplyEncoder {
	switch c {
	case JsonCodec:
		return codec.StdReplyEncoder
	case ProtobufCodec:
		return codec.ProtobufReplyEncoder
	default:
		gobase.AssertHere()
	}
	return nil
}
