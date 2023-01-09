package gorpc

import (
	"fmt"
	"github.com/BabySid/gorpc/grpc"
	"github.com/BabySid/gorpc/http"
	"net/url"
)

func DialHttpClient(rawUrl string, opts ...http.Option) (*http.Client, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "http", "https":
		return http.Dial(rawUrl, opts...)
	default:
		return nil, fmt.Errorf("no known transport for URL scheme %q", u.Scheme)
	}
}

func DialGRPCClient(rawUrl string) (*grpc.Client, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "grpc":
		return grpc.Dial(rawUrl)
	default:
		return nil, fmt.Errorf("no known transport for URL scheme %q", u.Scheme)
	}
}
