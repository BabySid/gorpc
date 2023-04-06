package main

import (
	"fmt"
	"github.com/BabySid/gorpc"
	"github.com/BabySid/gorpc/api"
	"github.com/BabySid/gorpc/codec"
)

func main() {
	testClient()
}

type SubData struct {
	DT string `json:"dt"`
}

func testClient() {
	recv := make(chan SubData)
	c, err := gorpc.Dial("ws://localhost:8888/_ws_", api.ClientOption{
		Codec:         codec.JsonCodec,
		WebSocketMode: api.WSM_JsonRpc,
		RevChan:       recv,
	})

	if err != nil {
		panic(any(err))
	}

	var param Params
	param.A = 100
	param.B = 200

	var res Result
	err = c.CallJsonRpc(&res, "rpc.Add", param)
	fmt.Println("Call rpc.Add return", res, err)

	var res2 Result2
	err = c.CallJsonRpc(&res2, "rpc.Add2", param)
	fmt.Println("Call rpc.Add2 return", res2, err.Error())
	apiErr, ok := api.FromError(err)
	fmt.Println(apiErr, ok)

	var res3 Result
	res3 = -1
	err = c.CallJsonRpc(&res3, "rpc.Add3", nil)
	fmt.Println("Call rpc.Add3 return", res3, err)

	var res4 string
	err = c.CallJsonRpc(&res4, "rpc.Sub", nil)
	fmt.Println("Call rpc.Sub return", res4, err)

	for {
		select {
		case data := <-recv:
			fmt.Println("rev from chan: ", data)
		}
	}
}

type Params struct {
	A int `json:"a"`
	B int `json:"b"`
}

type Result = int

type Result2 struct {
	C int `json:"c"`
}
