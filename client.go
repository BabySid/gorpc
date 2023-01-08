package gorpc

import (
	"fmt"
	"net/url"
)

type Client struct {
}

func Dial(rawUrl string) (*Client, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "http", "https":
		return nil, nil
	case "grpc":
		return nil, nil
	default:
		return nil, fmt.Errorf("no known transport for URL scheme %q", u.Scheme)
	}
}
