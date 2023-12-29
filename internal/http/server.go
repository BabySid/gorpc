package http

import (
	"errors"
	"fmt"
	"github.com/BabySid/gobase"
	"github.com/BabySid/gorpc/api"
	"github.com/BabySid/gorpc/internal/gin"
	"github.com/BabySid/gorpc/internal/jsonrpc"
	"github.com/BabySid/gorpc/internal/websocket"
	g "github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Server struct {
	opt        api.ServerOption
	httpServer *gin.Server

	rpcServer *jsonrpc.Server
}

func NewServer(option api.ServerOption) *Server {
	s := &Server{
		opt:        option,
		httpServer: gin.NewServer(),
		rpcServer:  jsonrpc.NewServer(jsonrpc.Option{CodeType: option.Codec}),
	}

	s.setUpBuiltInService()
	return s
}

func (s *Server) setUpBuiltInService() {
	s.httpServer.POST(api.BuiltInPathJsonRPC, s.processJsonRpc)
	s.httpServer.GET(api.BuiltInPathJsonWS, s.processJsonRpcByWS)

	if s.opt.EnableInnerService {
		s.httpServer.GET(api.BuiltInPathMetrics, g.WrapH(promhttp.Handler()))

		appName := filepath.Base(os.Args[0])
		indexHtml := fmt.Sprintf(`
<h2>WelCome to %s</h2>
<table border="1">
  <tr>
    <th>InnerPath</th>
  </tr>
  <tr>
    <td><a href="/%s">directory of %s</a></td>
  </tr>
  <tr>
	<td><a href="/%s">metrics of %s</a></td>
  </tr>
</table>
`, appName, api.BuiltInPathDIR, appName, api.BuiltInPathMetrics, appName)

		s.httpServer.GET("/", func(ctx *g.Context) {
			ctx.Header("Content-Type", "text/html; charset=utf-8")

			ctx.String(http.StatusOK, indexHtml)
		})
	}
}

func (s *Server) RegisterJsonRPC(name string, receiver interface{}) error {
	return s.rpcServer.RegisterName(name, receiver)
}

func (s *Server) RegisterPath(httpMethod string, path string, handle api.RawHandle) error {
	if err := s.checkPath(path); err != nil {
		return err
	}
	switch httpMethod {
	case http.MethodGet:
		s.httpServer.GET(path, getHandleWrapper(handle))
	case http.MethodPost:
		s.httpServer.POST(path, postHandleWrapper(handle))
	default:
		gobase.AssertHere()
	}
	return nil
}

var (
	invalidPath = errors.New("path is invalid. conflict with builtin")
)

func (s *Server) checkPath(path string) error {
	rootPath := ""
	if i := strings.Index(path, "/"); i >= 0 {
		if j := strings.Index(path[i+1:], "/"); j >= 0 {
			rootPath = path[i+1 : j+1]
		} else {
			rootPath = path[i+1:]
		}
	} else {
		rootPath = path
	}

	if rootPath == api.BuiltInPathMetrics ||
		rootPath == api.BuiltInPathJsonRPC ||
		rootPath == api.BuiltInPathJsonWS ||
		rootPath == api.BuiltInPathDIR {
		return invalidPath
	}

	return nil
}

func (s *Server) Run(ln net.Listener) error {
	path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}

	dir := http.Dir(path + "/..")
	log.Infof("set static fs to %s", dir)
	s.httpServer.StaticFS(api.BuiltInPathDIR, dir)

	return s.httpServer.RunListener(ln)
}

func (s *Server) processJsonRpcByWS(c *g.Context) {
	srv, err := websocket.NewServer(s.opt, s.rpcServer, c)
	if err != nil {
		c.String(http.StatusBadRequest, "websocket.NewServer: %s", err)
		return
	}
	defer srv.Close()
	srv.Run()
}

func (s *Server) processJsonRpc(c *g.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		resp := api.NewErrorJsonRpcResponse(nil, api.InternalError, api.SysCodeMap[api.InternalError], err.Error())
		c.JSON(http.StatusOK, resp)
		return
	}

	ctx := newHttpContext("jsonRpc2", uuid.New().String(), len(body), c)
	defer func() {
		ctx.EndRequest(api.Success)
	}()

	resp := s.rpcServer.Call(ctx, body)
	c.JSON(http.StatusOK, resp)
}
