package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/BabySid/gobase"
	"github.com/BabySid/gobase/log"
	"github.com/BabySid/gorpc"
	"github.com/BabySid/gorpc/api"
	"github.com/BabySid/gorpc/codec"
)

var l log.Logger

func main() {
	l = log.NewSLogger(log.WithOutFile("./server.log"), log.WithLevel("trace"))
	s := gorpc.NewServer(api.ServerOption{
		Addr:               ":8888",
		ClusterName:        "test",
		Logger:             l,
		JsonRpcOpt:         &api.JsonRpcOption{Codec: codec.JsonCodec},
		EnableInnerService: true,
	})

	t := &srv{}
	s.RegisterPath(http.MethodGet, "/v1/get/:uid", t.getHandle)
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

	l.Info("rawWsHandle processed",
		slog.String("notifier.ID", notifier.ID()), slog.Any("msgType", msg.Type), slog.String("msgData", string(msg.Data)), slog.Any("err", err))
	return nil
}

func (s *srv) getHandle(ctx api.RawHttpContext, httpBody []byte) {
	uid := ctx.Param("uid")
	name := ctx.Query("name")
	l.Info("getHandle", slog.String("uid", uid), slog.String("name", name))
	_ = ctx.WriteData(200, "text/plain; charset=utf-8", []byte("ok"))
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
	l.Info("Add3", slog.Any("params", params), slog.Any("result", result))

	return &result, nil
}

func (i *rpcServer) Add(ctx api.Context, params *Params) (*Result, *api.JsonRpcError) {
	a := params.A + params.B
	result := interface{}(a).(Result)
	l.Info("Add", slog.Any("result", result))
	return &result, nil
}

func (i *rpcServer) Add2(ctx api.Context, params *Params) (*Result2, *api.JsonRpcError) {
	var result Result2
	result.C = params.A + params.B
	l.Info("Add2", slog.Any("result", result))
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
				l.Warn("recv err", slog.Any("err", err))
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
