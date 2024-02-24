package api

import (
	"github.com/BabySid/gorpc/codec"
)

type Rotator struct {
	LogMaxAge int
	LogPath   string
}

type JsonRpcOption struct {
	Codec codec.CodecType
}
type ServerOption struct {
	Addr        string
	ClusterName string
	// logs
	Rotator  *Rotator
	LogLevel string

	JsonRpcOpt *JsonRpcOption

	BeforeRun          func() error
	EnableInnerService bool
}

const (
	BuiltInPathJsonRPC   = "_jsonrpc_"
	BuiltInPathWsJsonRPC = "_ws_"
	BuiltInPathRawWS     = "_raw_ws_"

	BuiltInPathDIR     = "_dir_"
	BuiltInPathMetrics = "_metrics_"
)

type BasicAuth struct {
	User   string
	Passwd string
}

type ClientOption struct {
	JsonRpcOpt *JsonRpcOption

	// http auth
	Basic *BasicAuth

	// websocket
	RevChan interface{}
}
