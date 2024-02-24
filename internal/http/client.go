package http

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/BabySid/gobase"
	"github.com/BabySid/gorpc/api"
	"github.com/BabySid/gorpc/internal/jsonrpc"
	"io"
	"net/http"
	"net/url"
)

var (
	ErrNoResult = errors.New("no result in JSON-RPC response")
)

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
		code, body, err := c.doPostHttp(c.rawUrl, reqs[0])
		if err != nil {
			return nil, err
		}
		defer body.Close()

		if err = c.checkHttpError(code, body); err != nil {
			return nil, err
		}

		data, err := io.ReadAll(body)
		if err != nil {
			return nil, err
		}

		var resp jsonrpc.Message
		if err = json.NewDecoder(bytes.NewReader(data)).Decode(&resp); err != nil {
			return nil, err
		}
		return []*jsonrpc.Message{&resp}, nil
	})
	return err
}

func (c *Client) BatchCallJsonRpc(b []api.BatchElem) error {
	gobase.True(c.jsonRpcCli != nil)
	err := c.jsonRpcCli.BatchCall(b, func(reqs ...*jsonrpc.Message) ([]*jsonrpc.Message, error) {
		gobase.True(len(reqs) > 0)
		code, body, err := c.doPostHttp(c.rawUrl, reqs)
		if err != nil {
			return nil, err
		}
		defer body.Close()

		if err = c.checkHttpError(code, body); err != nil {
			return nil, err
		}

		data, err := io.ReadAll(body)
		if err != nil {
			return nil, err
		}
		var resps []*jsonrpc.Message
		if err = json.NewDecoder(bytes.NewReader(data)).Decode(&resps); err != nil {
			return nil, err
		}
		return resps, nil
	})

	return err
}

func (c *Client) RawCallHttp(method string, path string, body interface{}) (int, io.ReadCloser, error) {
	switch method {
	case http.MethodGet:
		return c.doGetHttp(c.rawUrl + path)
	case http.MethodPost:
		return c.doPostHttp(c.rawUrl+path, body)
	default:
		gobase.AssertHere()
	}

	return 0, nil, nil
}

func Dial(rawUrl string, opt api.ClientOption) (*Client, error) {
	_, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	headers := http.Header{}
	headers.Set("accept", "application/json")
	headers.Set("content-type", "application/json")
	if opt.Basic != nil {
		auth := base64.StdEncoding.EncodeToString([]byte(opt.Basic.User + ":" + opt.Basic.Passwd))
		headers.Set("Authorization", "Basic "+auth)
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

func (c *Client) checkHttpError(code int, body io.ReadCloser) error {
	if code < http.StatusOK || code >= http.StatusMultipleChoices {
		bs, err := io.ReadAll(body)
		if err != nil {
			return err
		}
		return errors.New(fmt.Sprintf("%d:%s", code, string(bs)))
	}

	return nil
}

func (c *Client) doGetHttp(url string) (int, io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, nil, err
	}
	req.Header = c.header.Clone()

	resp, err := c.httpHandle.Do(req)
	if err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, resp.Body, nil
}

func (c *Client) doPostHttp(url string, msg interface{}) (int, io.ReadCloser, error) {
	body, err := json.Marshal(msg)
	if err != nil {
		return 0, nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, nil, err
	}
	req.ContentLength = int64(len(body))
	//req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(body)), nil }
	req.Header = c.header.Clone()

	resp, err := c.httpHandle.Do(req)
	if err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, resp.Body, nil
}
