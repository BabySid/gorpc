package websocket

import (
	"encoding/json"
	"fmt"
	"github.com/BabySid/gobase"
	"github.com/BabySid/gorpc/api"
	"github.com/BabySid/gorpc/internal/jsonrpc"
	"github.com/gorilla/websocket"
	"net/http"
	"reflect"
	"sync"
)

type Client struct {
	api.ClientAdapter

	rawUrl string
	opt    api.ClientOption

	conn       *websocket.Conn
	jsonRpcCli *jsonrpc.Client

	msgType reflect.Type
	msgChan reflect.Value

	errChan chan error
	close   chan struct{}

	respWait sync.Map
}

func (c *Client) Close() error {
	c.close <- struct{}{}
	return nil
}

func Dial(rawUrl string, opt api.ClientOption) (*Client, error) {
	chanVal := reflect.ValueOf(opt.RevChan)
	if chanVal.Kind() != reflect.Chan || chanVal.Type().ChanDir()&reflect.SendDir == 0 {
		panic(fmt.Sprintf("channel argument of Subscribe has type %T, need writable channel", opt.RevChan))
	}
	if chanVal.IsNil() {
		panic("channel given to Subscribe must not be nil")
	}

	dialer := websocket.Dialer{
		ReadBufferSize:  wsReadBuffer,
		WriteBufferSize: wsWriteBuffer,
		WriteBufferPool: wsBufferPool,
	}
	conn, resp, err := dialer.Dial(rawUrl, http.Header{})
	if err != nil {
		hErr := wsHandshakeError{err: err}
		if resp != nil {
			hErr.status = resp.Status
		}
		return nil, hErr
	}

	c := &Client{
		rawUrl:     rawUrl,
		opt:        opt,
		conn:       conn,
		jsonRpcCli: jsonrpc.NewClient(opt.Codec),
		msgType:    chanVal.Type().Elem(),
		msgChan:    chanVal,
		errChan:    make(chan error),
		close:      make(chan struct{}),
		respWait:   sync.Map{},
	}

	go c.read()

	return c, nil
}

type wsHandshakeError struct {
	err    error
	status string
}

func (e wsHandshakeError) Error() string {
	s := e.err.Error()
	if e.status != "" {
		s += " (HTTP status " + e.status + ")"
	}
	return s
}

func (c *Client) GetType() api.ClientType {
	return api.WsClient
}

func (c *Client) CallJsonRpc(result interface{}, method string, args interface{}) error {
	err := c.jsonRpcCli.Call(result, method, args, func(reqs ...*jsonrpc.Message) ([]byte, error) {
		gobase.True(len(reqs) == 1)
		ctx := rpcCallContext{
			id:   reqs[0].ID,
			resp: make(chan []byte),
		}
		c.respWait.Store(ctx.id, &ctx)
		err := c.WriteByWs(reqs[0])
		if err != nil {
			return nil, err
		}

		body := <-ctx.resp

		return body, nil
	})

	return err
}

func (c *Client) BatchCallJsonRpc(b []api.BatchElem) error {
	err := c.jsonRpcCli.BatchCall(b, func(reqs ...*jsonrpc.Message) ([]byte, error) {
		gobase.True(len(reqs) > 0)

		for _, req := range reqs {
			ctx := rpcCallContext{
				id:   req.ID,
				resp: make(chan []byte),
			}
			c.respWait.Store(ctx.id, &ctx)
		}

		err := c.WriteByWs(reqs)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	return err
}

func (c *Client) WriteByWs(req interface{}) error {
	bs, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return c.conn.WriteMessage(websocket.TextMessage, bs)
}

func (c *Client) read() {
	defer func() {
		c.Close()
	}()

	for {
		select {
		case <-c.close:
			return
		default:
			_, msg, err := c.conn.ReadMessage()
			if err != nil {
				c.errChan <- err
				return
			}
			// response of call? or subscription
			if c.opt.RawMode {
				err = c.handleRaw(msg)
			} else {
				err = c.handleJsonRpc(msg)
			}
			if err != nil {
				c.errChan <- err
				return
			}
		}
	}
}

func (c *Client) handleRaw(msg []byte) error {
	val := reflect.New(c.msgType)
	_ = json.Unmarshal(msg, val.Interface())
	c.msgChan.Send(reflect.ValueOf(val.Elem().Interface()))
	return nil
}

func (c *Client) handleJsonRpc(msg []byte) error {
	msgs, batch, err := jsonrpc.ParseBatchMessage(msg)
	if err != nil {
		return err
	}

	if batch {

	} else {
		ctx, ok := c.respWait.LoadAndDelete(msgs[0].ID)
		if ok {
			ctx.(*rpcCallContext).resp <- msg
		}
	}

	return nil
}

type rpcCallContext struct {
	id   interface{}
	resp chan []byte
}
