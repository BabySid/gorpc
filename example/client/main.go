package main

import (
	"fmt"
	"github.com/BabySid/gorpc"
	"github.com/BabySid/gorpc/api"
	"time"
)

func main() {
	testClient()
}

type SubData struct {
	DT string `json:"dt"`
}

func testClient() {
	recv := make(chan api.WSMessage)
	c, err := gorpc.Dial("ws://localhost:8888/_raw_ws_", api.ClientOption{
		RevChan: recv,
	})

	defer c.Close()

	if err != nil {
		panic(any(err))
	}

	go func() {
		for {
			select {
			case err := <-c.ErrFromWS():
				fmt.Println("err from ws: ", err)
				return
			case data := <-recv:
				fmt.Printf("rev from chan: %d %s\n", data.Type, string(data.Data))
			}
		}
	}()

	inputs := []string{
		"this is first",
		"this is 2",
		"3?",
	}
	for _, item := range inputs {
		err := c.WriteByWs(api.WSMessage{
			Type: api.WSTextMessage,
			Data: []byte(item),
		})
		fmt.Printf("send to server: %s %v\n", item, err)
		time.Sleep(time.Second * 3)
	}
	//var param Params
	//param.A = 100
	//param.B = 200
	//
	//var res Result
	//err = c.CallJsonRpc(&res, "rpc.Add", param)
	//fmt.Println(gobase.FormatDateTime(), " Call rpc.Add return", res, err)

	//var res2 Result2
	//err = c.CallJsonRpc(&res2, "rpc.Add2", param)
	//fmt.Println("Call rpc.Add2 return", res2, err.Error())
	//apiErr, ok := api.FromError(err)
	//fmt.Println(apiErr, ok)
	//
	//var res3 Result
	//res3 = -1
	//err = c.CallJsonRpc(&res3, "rpc.Add3", nil)
	//fmt.Println("Call rpc.Add3 return", res3, err)
	//
	//var res4 string
	//err = c.CallJsonRpc(&res4, "rpc.Sub", nil)
	//fmt.Println(gobase.FormatDateTime(), " Call rpc.Sub return", res4, err)
	//
	time.Sleep(30 * time.Second)

}

type Params struct {
	A int `json:"a"`
	B int `json:"b"`
}

type Result = int

type Result2 struct {
	C int `json:"c"`
}
