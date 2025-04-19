package api

import (
	"encoding/base64"
	"net/http"

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

func (ba BasicAuth) SetAuthHeader(head http.Header) {
	auth := base64.StdEncoding.EncodeToString([]byte(ba.User + ":" + ba.Passwd))
	head.Set("Authorization", "Basic "+auth)
}

type ClientOption struct {
	JsonRpcOpt *JsonRpcOption

	// http auth
	Heads http.Header

	// websocket
	RevChan interface{}
}

type WithHttpHeader func(http.Header)

func ResetHeader(key string, value string) WithHttpHeader {
	return func(head http.Header) {
		head.Set(key, value)
	}
}

func AppendHeader(key string, value string) WithHttpHeader {
	return func(head http.Header) {
		head.Add(key, value)
	}
}

var (
	WithContTypeAppJsonHeader = ResetHeader("Content-Type", "application/json")
	WithAcceptAppJsonHeader   = ResetHeader("Accept", "application/json")
)
