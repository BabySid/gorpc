package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/BabySid/gobase"
	"github.com/BabySid/gorpc/api"
	"github.com/BabySid/gorpc/internal/jsonrpc"
)

var ErrNoResult = errors.New("no result in JSON-RPC response")

type Client struct {
	api.ClientAdapter

	httpHandle *http.Client
	jsonRpcCli *jsonrpc.Client

	rawUrl string // e.g. https://localhost:8080/path
	opt    api.ClientOption
	header http.Header
}

func (c *Client) GetType() api.ClientType {
	return api.HttpClient
}

func (c *Client) CallJsonRpc(result interface{}, method string, args interface{}) error {
	gobase.True(c.jsonRpcCli != nil)
	err := c.jsonRpcCli.Call(result, method, args, func(reqs ...*jsonrpc.Message) ([]*jsonrpc.Message, error) {
		gobase.True(len(reqs) == 1)
		resp, err := c.doPostHttp(c.rawUrl, reqs[0], api.WithAcceptAppJsonHeader, api.WithContTypeAppJsonHeader)
		if err != nil {
			return nil, err
		}

		if err = c.checkHttpError(resp); err != nil {
			return nil, err
		}

		var res jsonrpc.Message
		if err = json.NewDecoder(bytes.NewReader(resp.Body)).Decode(&res); err != nil {
			return nil, err
		}
		return []*jsonrpc.Message{&res}, nil
	})
	return err
}

func (c *Client) BatchCallJsonRpc(b []api.BatchElem) error {
	gobase.True(c.jsonRpcCli != nil)
	err := c.jsonRpcCli.BatchCall(b, func(reqs ...*jsonrpc.Message) ([]*jsonrpc.Message, error) {
		gobase.True(len(reqs) > 0)
		resp, err := c.doPostHttp(c.rawUrl, reqs, api.WithAcceptAppJsonHeader, api.WithContTypeAppJsonHeader)
		if err != nil {
			return nil, err
		}

		if err = c.checkHttpError(resp); err != nil {
			return nil, err
		}

		var resps []*jsonrpc.Message
		if err = json.NewDecoder(bytes.NewReader(resp.Body)).Decode(&resps); err != nil {
			return nil, err
		}
		return resps, nil
	})

	return err
}

func (c *Client) RawCallHttp(method string, path string, body interface{}) (*api.HttpResponse, error) {
	switch method {
	case http.MethodGet:
		return c.doGetHttp(c.rawUrl + path)
	case http.MethodPost:
		return c.doPostHttp(c.rawUrl+path, body)
	default:
		gobase.AssertHere()
	}
	return nil, nil
}

func Dial(rawUrl string, opt api.ClientOption) (*Client, error) {
	_, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	headers := http.Header{}

	for key, values := range opt.Heads {
		for _, value := range values {
			headers.Add(key, value)
		}
	}

	c := &Client{
		rawUrl:     rawUrl,
		httpHandle: new(http.Client),
		jsonRpcCli: nil,
		header:     headers,
		opt:        opt,
	}

	if opt.JsonRpcOpt != nil {
		c.jsonRpcCli = jsonrpc.NewClient(opt.JsonRpcOpt.Codec)
	}

	return c, nil
}

func (c *Client) checkHttpError(resp *api.HttpResponse) error {
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return errors.New(fmt.Sprintf("%d:%s", resp.StatusCode, string(resp.Body)))
	}

	return nil
}

func (c *Client) doGetHttp(url string, opts ...api.WithHttpHeader) (*api.HttpResponse, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header = c.header.Clone()
	for _, opt := range opts {
		opt(req.Header)
	}

	resp, err := c.httpHandle.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	res := api.HttpResponse{}
	res.StatusCode = resp.StatusCode
	res.Body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *Client) doPostHttp(url string, msg any, opts ...api.WithHttpHeader) (*api.HttpResponse, error) {
	body, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.ContentLength = int64(len(body))
	// req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(body)), nil }
	req.Header = c.header.Clone()
	for _, opt := range opts {
		opt(req.Header)
	}

	resp, err := c.httpHandle.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	res := api.HttpResponse{}
	res.StatusCode = resp.StatusCode
	res.Body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) Close() error { return nil }
