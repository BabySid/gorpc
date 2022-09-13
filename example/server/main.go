package main

import (
	"encoding/json"
	"github.com/BabySid/gorpc"
	"github.com/BabySid/gorpc/http/httpapi"
	"net/http"
)

func main() {
	s := gorpc.NewServer()

	t := &srv{}
	s.RegisterPath(http.MethodGet, "/v1/get", t.getHandle)
	s.RegisterPath(http.MethodPost, "/v1/post", t.postHandle)

	_ = s.Run(gorpc.ServerOption{
		Addr:        ":8888",
		ClusterName: "test",
		Rotator:     nil,
		LogLevel:    "info",
	})
}

type srv struct{}

func (s *srv) getHandle(ctx *httpapi.APIContext, httpBody *json.RawMessage) (interface{}, *httpapi.JsonRpcError) {
	return "hello world", nil
}

func (s *srv) postHandle(ctx *httpapi.APIContext, httpBody *json.RawMessage) (interface{}, *httpapi.JsonRpcError) {
	return map[string]interface{}{
		"hello": 123,
	}, nil
}
