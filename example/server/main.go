package main

import (
	"errors"
	"fmt"
	"github.com/BabySid/gorpc"
	"github.com/BabySid/gorpc/api"
	"github.com/BabySid/gorpc/codec"
	"github.com/BabySid/proto/sodor"
	"net/http"
	"time"
)

func main() {
	s := gorpc.NewServer(api.ServerOption{
		Addr:        ":8888",
		ClusterName: "test",
		Rotator:     nil,
		LogLevel:    "info",
		Codec:       codec.JsonCodec,
	})

	t := &srv{}
	s.RegisterPath(http.MethodGet, "/v1/get", t.getHandle)
	s.RegisterPath(http.MethodPost, "/v1/post", t.postHandle)
	s.RegisterJsonRPC("rpc", &rpcServer{})

	go func() {
		time.Sleep(5 * time.Second)
		fmt.Println("begin run Client...")
		testClient()
	}()
	err := s.Run()
	fmt.Println(err)
}

func testSodor() {
	c, err := gorpc.Dial("http://172.28.17.127:9527", api.ClientOption{Codec: codec.ProtobufCodec})
	if err != nil {
		panic(any(err))
	}

	resp := sodor.ThomasInstance{}
	req := sodor.ThomasInfo{}
	req.Id = 2
	err = c.CallJsonRpc(&resp, "rpc.ShowThomas", &req)
	fmt.Println(err)
	fmt.Println(resp)
}

func testClient() {
	c, err := gorpc.Dial("http://localhost:8888/_jsonrpc_", api.ClientOption{Codec: codec.JsonCodec})
	if err != nil {
		panic(any(err))
	}

	//var param Params
	//param.A = 100
	//param.B = 200
	//
	//var res Result
	//err = c.CallJsonRpc(&res, "rpc.Add", param)
	//fmt.Println("Call rpc.Add return", res, err)
	//
	//var res2 Result2
	//err = c.CallJsonRpc(&res2, "rpc.Add2", param)
	//fmt.Println("Call rpc.Add2 return", res2, err.Error())
	//apiErr, ok := api.FromError(err)
	//fmt.Println(apiErr, ok)

	var res3 Result
	res3 = -1
	err = c.CallJsonRpc(&res3, "rpc.Add3", nil)
	fmt.Println("Call rpc.Add3 return", res3, err)

	//code, body, err := c.RawCallHttp(http.MethodGet, "/v1/get", nil)
	//data, _ := ioutil.ReadAll(body)
	//
	//fmt.Println("RawCall Get return", code, string(data), err)
	//
	//code, body, err = c.RawCallHttp(http.MethodPost, "/v1/post", map[string]interface{}{
	//	"id":     123,
	//	"params": "hello rpc",
	//})
	//data, _ = ioutil.ReadAll(body)
	//
	//fmt.Println("RawCall Post return", code, string(data), err)
}

type srv struct{}

func (s *srv) getHandle(ctx api.Context, httpBody interface{}) *api.JsonRpcResponse {
	return api.NewSuccessJsonRpcResponse(ctx.ID(), "hello world")
}

func (s *srv) postHandle(ctx api.Context, httpBody interface{}) *api.JsonRpcResponse {
	if httpBody != nil {
		ctx.Log("httpBody %v", httpBody)
	}

	return api.NewSuccessJsonRpcResponse(ctx.ID(), map[string]interface{}{
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

func (i *rpcServer) Add3(ctx api.Context, params interface{}) (*Result, *api.JsonRpcError) {
	a := 100
	result := interface{}(a).(Result)
	ctx.Log("Add3 %v params=%v", result, params)
	return &result, nil
}

func (i *rpcServer) Add(ctx api.Context, params *Params) (*Result, *api.JsonRpcError) {
	a := params.A + params.B
	result := interface{}(a).(Result)
	ctx.Log("Add %v", result)
	return &result, nil
}

func (i *rpcServer) Add2(ctx api.Context, params *Params) (*Result2, *api.JsonRpcError) {
	var result Result2
	result.C = params.A + params.B
	ctx.Log("Add2 %v", result)
	if result.C%100 == 0 {
		return nil, api.NewJsonRpcError(-32000, "bad param", errors.New("aha error"))
	}
	return &result, nil
}
