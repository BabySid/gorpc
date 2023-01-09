package grpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/url"
)

type Client struct {
	*grpc.ClientConn
}

func Dial(rawUrl string) (*Client, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	target := u.Hostname() + u.Port()
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	c := Client{conn}
	return &c, err
}
