package api

import (
	"github.com/BabySid/gobase/log"
	"github.com/BabySid/gorpc/codec"
)

type JsonRpcOption struct {
	Codec codec.CodecType
}
type ServerOption struct {
	Addr        string
	ClusterName string

	Logger log.Logger

	JsonRpcOpt *JsonRpcOption

	BeforeRun          func() error
	EnableInnerService bool
}

const (
	BuiltInPathJsonRPC   = "_jsonrpc_"
	BuiltInPathWsJsonRPC = "_jsonrpc_ws_"
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
