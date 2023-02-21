package api

import (
	"github.com/BabySid/gorpc/codec"
)

type Rotator struct {
	LogMaxAge int
	LogPath   string
}

type ServerOption struct {
	Addr        string
	ClusterName string
	// logs
	Rotator  *Rotator
	LogLevel string

	Codec codec.CodecType

	BeforeRun func() error
}

const (
	BuiltInPathJsonRPC = "_jsonrpc_"
	BuiltInPathJsonWS  = "_ws_"
	BuiltInPathDIR     = "_dir_"
	BuiltInPathMetrics = "_metrics_"
)

type ClientOption struct {
	Codec codec.CodecType

	// websocket
	RevChan interface{}
	// return raw response for subscription because proto of response on some server is not jsonrpc
	WebSocketMode WebSocketMode
}

type WebSocketMode int

const (
	WSM_RawJson = iota
	WSM_JsonRpc
)
