package gorpc

import (
	"fmt"
	"github.com/BabySid/gorpc/api"
	"github.com/BabySid/gorpc/internal/grpc"
	"github.com/BabySid/gorpc/internal/http"
	"github.com/BabySid/gorpc/internal/websocket"
	"net/url"
)

func Dial(rawUrl string, opt api.ClientOption) (api.Client, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "http", "https":
		return http.Dial(rawUrl, opt)
	case "ws", "wss":
		return websocket.Dial(rawUrl, opt)
	case "grpc":
		return grpc.Dial(rawUrl)
	default:
		return nil, fmt.Errorf("no known transport for URL scheme %q", u.Scheme)
	}
}
