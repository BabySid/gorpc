package httpcfg

import "github.com/BabySid/gorpc/http/codec"

type ParamsDecoder func(in interface{}, out interface{}) error
type ServerOption struct {
	PDecoder ParamsDecoder
}

var (
	StdParamsDecoder      = codec.StdParamsDecoder
	ProtoBufParamsDecoder = codec.ProtobufParamsDecoder
	DefaultOption         = ServerOption{PDecoder: StdParamsDecoder}
)
