package main

import (
	"errors"
	"fmt"
	"github.com/BabySid/gorpc"
	"github.com/BabySid/gorpc/http/httpapi"
	"github.com/BabySid/gorpc/http/httpcfg"
	"net/http"
)

func main() {
	s := gorpc.NewServer(gorpc.ServerOption{
		Addr:        ":8888",
		ClusterName: "test",
		Rotator:     nil,
		LogLevel:    "info",
		HttpOpt:     httpcfg.DefaultOption,
	})

	t := &srv{}
	s.RegisterPath(http.MethodGet, "/v1/get", t.getHandle)
	s.RegisterPath(http.MethodPost, "/v1/post", t.postHandle)
	s.RegisterJsonRPC("rpc", &rpcServer{})

	err := s.Run()
	fmt.Println(err)
}

type srv struct{}

func (s *srv) getHandle(ctx *httpapi.APIContext, httpBody interface{}) *httpapi.JsonRpcResponse {
	return httpapi.NewSuccessJsonRpcResponse(ctx.ID, "hello world")
}

func (s *srv) postHandle(ctx *httpapi.APIContext, httpBody interface{}) *httpapi.JsonRpcResponse {
	if httpBody != nil {
		ctx.ToLog("httpBody %v", httpBody)
	}

	return httpapi.NewSuccessJsonRpcResponse(ctx.ID, map[string]interface{}{
		"hello": httpBody,
	})
}

type rpcServer struct{}

type Params struct {
	A int `json:"a"`
	B int `json:"b"`
}

type Result = int

type Result2 struct {
	C int `json:"c"`
}

func (i *rpcServer) Add3(ctx *httpapi.APIContext, params *interface{}) (*Result, *httpapi.JsonRpcError) {
	a := 100
	result := interface{}(a).(Result)
	ctx.ToLog("Add %v", result)
	return &result, nil
}

func (i *rpcServer) Add(ctx *httpapi.APIContext, params *Params) (*Result, *httpapi.JsonRpcError) {
	a := params.A + params.B
	result := interface{}(a).(Result)
	ctx.ToLog("Add %v", result)
	return &result, nil
}

func (i *rpcServer) Add2(ctx *httpapi.APIContext, params *Params) (*Result2, *httpapi.JsonRpcError) {
	var result Result2
	result.C = params.A + params.B
	ctx.ToLog("Add2 %v", result)
	if result.C%100 == 0 {
		return nil, httpapi.NewJsonRpcError(-32000, "bad param", errors.New("aha error"))
	}
	return &result, nil
}
