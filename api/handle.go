package api

// RpcHandle is a raw interface for creating api based http
type RpcHandle func(ctx Context, params interface{}) *JsonRpcResponse
