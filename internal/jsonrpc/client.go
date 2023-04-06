package jsonrpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/BabySid/gobase"
	"github.com/BabySid/gorpc/api"
	"github.com/BabySid/gorpc/codec"
	"reflect"
	"strconv"
	"sync/atomic"
)

type Client struct {
	ct        codec.CodecType
	idCounter uint32
}

func NewClient(ct codec.CodecType) *Client {
	c := Client{ct: ct}

	return &c
}

type MessageReader func(reqs ...*Message) ([]*Message, error)

var (
	ErrNoResult = errors.New("no result in JSON-RPC response")
)

func (c *Client) Call(result interface{}, method string, args interface{}, reader MessageReader) error {
	if result != nil && reflect.TypeOf(result).Kind() != reflect.Ptr {
		return fmt.Errorf("call result parameter must be pointer or nil interface: %v", result)
	}
	msg, err := c.newMessage(method, args)
	if err != nil {
		return err
	}

	resps, err := reader(msg)
	if err != nil {
		return err
	}

	gobase.True(len(resps) == 1)

	if resps[0].Error != nil {
		return resps[0].Error
	}
	if len(resps[0].Result) == 0 {
		return ErrNoResult
	}

	switch c.ct {
	case codec.JsonCodec:
		return json.Unmarshal(resps[0].Result, result)
	case codec.ProtobufCodec:
		return codec.DefaultProtoMarshal.Unmarshal(resps[0].Result, result)
	default:
		gobase.AssertHere()
	}
	return nil
}

func (c *Client) BatchCall(batch []api.BatchElem, reader MessageReader) error {
	var (
		msgs = make([]*Message, len(batch))
		byID = make(map[string]int, len(batch))
	)

	for i, elem := range batch {
		msg, err := c.newMessage(elem.Method, elem.Args)
		if err != nil {
			return err
		}
		msgs[i] = msg
		byID[string(msg.ID)] = i
	}

	resps, err := reader(msgs...)
	if err != nil {
		return err
	}

	for n := 0; n < len(resps) && err == nil; n++ {
		res := resps[n]
		elem := &batch[byID[string(res.ID)]]
		if res.Error != nil {
			elem.Error = res.Error
			continue
		}
		if len(res.Result) == 0 {
			elem.Error = ErrNoResult
			continue
		}
		switch c.ct {
		case codec.JsonCodec:
			elem.Error = json.Unmarshal(res.Result, elem.Result)
		case codec.ProtobufCodec:
			elem.Error = codec.DefaultProtoMarshal.Unmarshal(res.Result, elem.Result)
		default:
			gobase.AssertHere()
		}
	}

	return err
}

func (c *Client) nextID() json.RawMessage {
	id := atomic.AddUint32(&c.idCounter, 1)
	return strconv.AppendUint(nil, uint64(id), 10)
}

func (c *Client) newMessage(method string, paramsIn interface{}) (*Message, error) {
	msg := &Message{Version: api.Version, ID: c.nextID(), Method: method}
	if paramsIn != nil { // prevent sending "params":null
		var err error
		switch c.ct {
		case codec.JsonCodec:
			if msg.Params, err = json.Marshal(paramsIn); err != nil {
				return nil, err
			}
		case codec.ProtobufCodec:
			if msg.Params, err = codec.DefaultProtoMarshal.Marshal(paramsIn); err != nil {
				return nil, err
			}
		default:
			gobase.AssertHere()
		}
	}
	return msg, nil
}
