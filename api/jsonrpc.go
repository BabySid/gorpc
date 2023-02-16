package api

import (
	"encoding/json"
	"fmt"
	"github.com/BabySid/gobase"
	"github.com/BabySid/gorpc/codec"
	"google.golang.org/protobuf/proto"
)

type JsonRpcRequest struct {
	Version string `json:"jsonrpc"`
	Method  string `json:"method"`
	// Params should be json.RawMessage. However, in fact, since the params parameter in the request body is already
	// an interface after the data is deserialized into a map, if it is set to json.rawMessage,
	// essentially a conversion process is required. In other words, the type of the value in map after deserialization
	// is interface{}, and the interface{} cannot be forced to be converted to json.RawMessage
	Params interface{} `json:"params,omitempty"`
	Id     interface{} `json:"id"`
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
	return fmt.Sprintf("jsonError(code: %d, message: %s, data: %v)", j.Code, j.Message, j.Data)
}

func NewJsonRpcErr(c int, data interface{}) *JsonRpcError {
	msg := SysCodeMap[c]
	gobase.True(msg != "")
	return NewJsonRpcError(c, msg, data)
}

func FromError(err error) (*JsonRpcError, bool) {
	if err == nil {
		return nil, true
	}
	if rpcErr, ok := err.(*JsonRpcError); ok {
		return rpcErr, true
	}

	return nil, false
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
