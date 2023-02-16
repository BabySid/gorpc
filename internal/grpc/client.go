package grpc

import (
	"github.com/BabySid/gorpc/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/url"
)

type Client struct {
	api.ClientAdapter

	*grpc.ClientConn
}

func (c *Client) GetType() api.ClientType {
	return api.GrpcClient
}

func (c *Client) Close() error {
	return c.ClientConn.Close()
}

func (c *Client) UnderlyingHandle() interface{} {
	return c.ClientConn
}

func Dial(rawUrl string) (*Client, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	target := u.Hostname() + u.Port()
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	c := Client{ClientConn: conn}
	return &c, err
}
