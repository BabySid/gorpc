package httpapi

import (
	"encoding/json"
	"fmt"
	"github.com/BabySid/gorpc/http/codec"
	"google.golang.org/protobuf/proto"
)

type JsonRpcRequest struct {
	Version string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	Id      interface{} `json:"id"`
}

type JsonRpcResponse struct {
	Version string          `json:"jsonrpc"`
	Id      interface{}     `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JsonRpcError   `json:"error,omitempty"`
}

type JsonRpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

const (
	Version = "2.0"
)

func (j *JsonRpcError) Error() string {
	return fmt.Sprintf("jsonError(code: %d, message: %s)", j.Code, j.Message)
}

func NewJsonRpcError(c int, msg string, data interface{}) *JsonRpcError {
	// wrapper for typeof error
	// otherwise user need call this func via `NewJsonRpcError(..., errors.New(...).String())`
	if err, ok := data.(error); ok {
		data = err.Error()
	}
	return &JsonRpcError{
		Code:    c,
		Message: msg,
		Data:    data,
	}
}

func NewSuccessJsonRpcResponse(id interface{}, result interface{}) *JsonRpcResponse {
	var rs json.RawMessage
	var err error
	if msg, ok := result.(proto.Message); ok {
		rs, err = codec.DefaultProtoMarshal.Marshal(msg)
	} else {
		rs, err = codec.StdReplyEncoder(result)
	}
	if err != nil {
		return nil
	}
	resp := &JsonRpcResponse{Version: Version, Id: id, Result: rs}
	return resp
}

func NewErrorJsonRpcResponse(id interface{}, code int, msg string, data interface{}) *JsonRpcResponse {
	err := &JsonRpcError{
		Code:    code,
		Message: msg,
		Data:    data,
	}

	return NewErrorJsonRpcResponseWithError(id, err)
}

func NewErrorJsonRpcResponseWithError(id interface{}, err *JsonRpcError) *JsonRpcResponse {
	resp := &JsonRpcResponse{Version: Version, Id: id}
	resp.Error = err

	return resp
}

// RpcHandle is a raw interface for creating api based http
type RpcHandle func(ctx *APIContext, params interface{}) *JsonRpcResponse
