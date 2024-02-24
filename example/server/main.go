package main

import (
	"errors"
	"fmt"
	"github.com/BabySid/gobase"
	"github.com/BabySid/gorpc"
	"github.com/BabySid/gorpc/api"
	"github.com/BabySid/gorpc/codec"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func main() {
	s := gorpc.NewServer(api.ServerOption{
		Addr:               ":8888",
		ClusterName:        "test",
		Rotator:            nil,
		LogLevel:           "trace",
		JsonRpcOpt:         &api.JsonRpcOption{Codec: codec.JsonCodec},
		EnableInnerService: true,
	})

	t := &srv{}
	s.RegisterPath(http.MethodGet, "/v1/get/:uid", t.getHandle)
	//s.RegisterPath(http.MethodPost, "/v1/post", t.postHandle)
	s.RegisterJsonRPC("rpc", &rpcServer{})
	s.RegisterRawWs(t.rawWsHandle)

	err := s.Run()
	fmt.Println(err)
}

type srv struct{}

func (s *srv) rawWsHandle(ctx api.Context, msg api.WSMessage) error {
	value, ok := ctx.Value(api.RawWSNotifierKey)
	if !ok {
		panic(false)
	}
	notifier, ok := value.(api.RawWSNotifier)
	if !ok {
		panic("panic if !ok")
	}

	tmp := string(msg.Data) + gobase.FormatDateTime()
	err := notifier.Write(api.WSMessage{
		Type: api.WSTextMessage,
		Data: []byte(tmp),
	})

	ctx.Log("%s => %d %s %v", notifier.ID(), msg.Type, string(msg.Data), err)
	return nil
}

func (s *srv) getHandle(ctx api.RawContext, httpBody []byte) {
	uid := ctx.Param("uid")
	name := ctx.Query("name")
	ctx.Log("got uid = %s name = %s", uid, name)
	_ = ctx.WriteData(200, "text/plain; charset=utf-8", []byte("ok"))
}

func (s *srv) postHandle(ctx api.Context, httpBody interface{}) *api.JsonRpcResponse {
	if httpBody != nil {
		ctx.Log("httpBody %v", httpBody)
	}

	return api.NewSuccessJsonRpcResponse(ctx.CtxID(), map[string]interface{}{
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
	_, ok := ctx.Value(api.JsonRpcNotifierKey)
	if ok {
		return nil, api.NewJsonRpcError(-32000, "not supported", errors.New("not supported"))
	}
	a := 100
	result := interface{}(a).(Result)
	ctx.Log("Add3 %v params=%v", result, params)
	return &result, nil
}

func (i *rpcServer) Add(ctx api.Context, params *Params) (*Result, *api.JsonRpcError) {
	a := params.A + params.B
	result := interface{}(a).(Result)
	//ctx.Log("Add %v err=%v", result, err)
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

type SubResult string

type SubData struct {
	DT string `json:"dt"`
}

func (i *rpcServer) Sub(ctx api.Context, params *Params) (*SubResult, *api.JsonRpcError) {
	value, ok := ctx.Value(api.JsonRpcNotifierKey)
	if !ok {
		return nil, api.NewJsonRpcError(-32000, "not supported", errors.New("not supported"))
	}
	notifier, ok := value.(api.JsonRpcNotifier)
	if !ok {
		panic("panic if !ok")
	}

	go func() {
		for {
			select {
			case err := <-notifier.Err():
				log.Infof("err found: %v", err)
				return
			default:
				time.Sleep(3 * time.Second)
				notifier.Notify(api.NewSubscriptionNotice("Sub", "0x5e7c550061dad01c4f59eab18b2e055", SubData{DT: gobase.FormatDateTime()}))
			}

		}

	}()

	var rs SubResult
	rs = "0x5e7c550061dad01c4f59eab18b2e055"

	return &rs, nil
}
