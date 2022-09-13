package httpapi

import (
	"encoding/json"
	"fmt"
)

type JsonRpcRequest struct {
	Method string           `json:"method"`
	Params *json.RawMessage `json:"params,omitempty"`
	Id     interface{}      `json:"id"`
}

type JsonRpcResponse struct {
	Id     interface{}   `json:"id"`
	Result interface{}   `json:"result,omitempty"`
	Error  *JsonRpcError `json:"error,omitempty"`
}

type JsonRpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (j *JsonRpcError) Error() string {
	return fmt.Sprintf("jsonError(code: %d, message: %s)", j.Code, j.Message)
}

func NewSuccessJsonRpcResponse(id interface{}, result interface{}) *JsonRpcResponse {
	resp := &JsonRpcResponse{Id: id, Result: result}
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
	resp := &JsonRpcResponse{}
	resp.Id = id
	resp.Error = err

	return resp
}

// RpcHandle is a raw interface for creating api based http
type RpcHandle func(ctx *APIContext, httpBody *json.RawMessage) (interface{}, *JsonRpcError)
