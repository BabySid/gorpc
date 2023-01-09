package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/BabySid/gobase"
	"github.com/BabySid/gorpc/http/codec"
	"github.com/BabySid/gorpc/http/httpapi"
	"github.com/BabySid/gorpc/http/httpcfg"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"sync/atomic"
)

type Option func(*Client)

func WithProtobufCodec() Option {
	return func(c *Client) {
		c.codec = httpcfg.ProtobufCodec
	}
}

// BatchElem is an element in a batch request.
type BatchElem struct {
	Method string
	Args   interface{}
	// The result is unmarshaled into this field. Result must be set to a
	// non-nil pointer value of the desired type, otherwise the response will be
	// discarded.
	Result interface{}
	// Error is set if the server returns an error for this request, or if
	// unmarshaling into Result fails. It is not set for I/O errors.
	Error error
}

type Client struct {
	endpoint  string
	handle    *http.Client
	headers   http.Header
	idCounter uint32
	codec     httpcfg.CodecType
}

var (
	ErrNoResult = errors.New("no result in JSON-RPC response")
)

func Dial(rawUrl string, opts ...Option) (*Client, error) {
	_, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	headers := http.Header{}
	headers.Set("accept", "application/json")
	headers.Set("content-type", "application/json")

	c := &Client{
		endpoint:  rawUrl,
		handle:    new(http.Client),
		headers:   headers,
		idCounter: 0,
		codec:     httpcfg.JsonCodec,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

func (c *Client) Call(result interface{}, method string, args interface{}) error {
	if result != nil && reflect.TypeOf(result).Kind() != reflect.Ptr {
		return fmt.Errorf("call result parameter must be pointer or nil interface: %v", result)
	}
	msg, err := c.newMessage(method, args)
	if err != nil {
		return err
	}

	code, body, err := c.doPostHttp("", msg)
	if err != nil {
		return err
	}
	defer body.Close()

	if err = c.checkHttpError(code, body); err != nil {
		return err
	}

	var resp jsonrpcMessage
	if err = json.NewDecoder(body).Decode(&resp); err != nil {
		return err
	}

	if resp.Error != nil {
		return resp.Error
	}
	if len(resp.Result) == 0 {
		return ErrNoResult
	}
	return json.Unmarshal(resp.Result, &result)
}

func (c *Client) BatchCall(b []BatchElem) error {
	var (
		msgs = make([]*jsonrpcMessage, len(b))
		byID = make(map[string]int, len(b))
	)

	for i, elem := range b {
		msg, err := c.newMessage(elem.Method, elem.Args)
		if err != nil {
			return err
		}
		msgs[i] = msg
		byID[string(msg.ID)] = i
	}

	code, body, err := c.doPostHttp("", msgs)
	if err != nil {
		return err
	}
	defer body.Close()

	if err = c.checkHttpError(code, body); err != nil {
		return err
	}

	var resp []jsonrpcMessage
	if err = json.NewDecoder(body).Decode(&resp); err != nil {
		return err
	}

	for n := 0; n < len(resp) && err == nil; n++ {
		res := resp[n]
		elem := &b[byID[string(res.ID)]]
		if res.Error != nil {
			elem.Error = res.Error
			continue
		}
		if len(res.Result) == 0 {
			elem.Error = ErrNoResult
			continue
		}
		elem.Error = json.Unmarshal(res.Result, elem.Result)
	}

	return err
}

func (c *Client) RawCall(method string, path string, body interface{}) (int, io.ReadCloser, error) {
	switch method {
	case http.MethodGet:
		return c.doGetHttp(path)
	case http.MethodPost:
		return c.doPostHttp(path, body)
	default:
		gobase.AssertHere()
	}

	return 0, nil, nil
}

func (c *Client) nextID() json.RawMessage {
	id := atomic.AddUint32(&c.idCounter, 1)
	return strconv.AppendUint(nil, uint64(id), 10)
}

func (c *Client) checkHttpError(code int, body io.ReadCloser) error {
	if code < http.StatusOK || code >= http.StatusMultipleChoices {
		bs, err := ioutil.ReadAll(body)
		if err != nil {
			return err
		}
		return errors.New(fmt.Sprintf("%d:%s", code, string(bs)))
	}

	return nil
}

func (c *Client) newMessage(method string, paramsIn interface{}) (*jsonrpcMessage, error) {
	msg := &jsonrpcMessage{Version: httpapi.Version, ID: c.nextID(), Method: method}
	if paramsIn != nil { // prevent sending "params":null
		var err error
		switch c.codec {
		case httpcfg.JsonCodec:
			if msg.Params, err = json.Marshal(paramsIn); err != nil {
				return nil, err
			}
		case httpcfg.ProtobufCodec:
			if msg.Params, err = codec.DefaultProtoMarshal.Marshal(paramsIn); err != nil {
				return nil, err
			}
		default:
			gobase.AssertHere()
		}
	}
	return msg, nil
}

func (c *Client) doGetHttp(path string) (int, io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, c.endpoint+path, nil)
	if err != nil {
		return 0, nil, err
	}
	req.Header = c.headers.Clone()

	resp, err := c.handle.Do(req)
	if err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, resp.Body, nil
}

func (c *Client) doPostHttp(path string, msg interface{}) (int, io.ReadCloser, error) {
	body, err := json.Marshal(msg)
	if err != nil {
		return 0, nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.endpoint+path, io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		return 0, nil, err
	}
	req.ContentLength = int64(len(body))
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(body)), nil }
	req.Header = c.headers.Clone()

	resp, err := c.handle.Do(req)
	if err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, resp.Body, nil
}

// A value of this type can a JSON-RPC request, notification, successful response or
// error response. Which one it is depends on the fields.
type jsonrpcMessage struct {
	Version string                `json:"jsonrpc,omitempty"`
	ID      json.RawMessage       `json:"id,omitempty"`
	Method  string                `json:"method,omitempty"`
	Params  json.RawMessage       `json:"params,omitempty"`
	Error   *httpapi.JsonRpcError `json:"error,omitempty"`
	Result  json.RawMessage       `json:"result,omitempty"`
}
