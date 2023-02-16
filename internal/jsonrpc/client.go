package jsonrpc

import (
	"bytes"
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

type MessageReader func(req interface{}) ([]byte, error)

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

	data, err := reader(msg)
	if err != nil {
		return err
	}

	var resp Message
	if err = json.NewDecoder(bytes.NewReader(data)).Decode(&resp); err != nil {
		return err
	}

	if resp.Error != nil {
		return resp.Error
	}
	if len(resp.Result) == 0 {
		return ErrNoResult
	}

	switch c.ct {
	case codec.JsonCodec:
		return json.Unmarshal(resp.Result, result)
	case codec.ProtobufCodec:
		return codec.DefaultProtoMarshal.Unmarshal(resp.Result, result)
	default:
		gobase.AssertHere()
	}
	return nil
}

func (c *Client) BatchCall(b []api.BatchElem, reader MessageReader) error {
	var (
		msgs = make([]*Message, len(b))
		byID = make(map[interface{}]int, len(b))
	)

	for i, elem := range b {
		msg, err := c.newMessage(elem.Method, elem.Args)
		if err != nil {
			return err
		}
		msgs[i] = msg
		byID[msg.ID] = i
	}

	data, err := reader(msgs)
	if err != nil {
		return err
	}

	var resp []Message
	if err = json.NewDecoder(bytes.NewReader(data)).Decode(&resp); err != nil {
		return err
	}

	for n := 0; n < len(resp) && err == nil; n++ {
		res := resp[n]
		elem := &b[byID[res.ID]]
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
