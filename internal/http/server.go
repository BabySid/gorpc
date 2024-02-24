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

	rawWsHandle api.RawWsHandle
}

func NewServer(option api.ServerOption) *Server {
	s := &Server{
		opt:        option,
		httpServer: gin.NewServer(),
		rpcServer:  nil,
	}

	if s.opt.JsonRpcOpt != nil {
		s.rpcServer = jsonrpc.NewServer(jsonrpc.Option{CodeType: s.opt.JsonRpcOpt.Codec})
	}

	s.setUpBuiltInService()
	return s
}

func (s *Server) setUpBuiltInService() {
	s.httpServer.POST(api.BuiltInPathJsonRPC, s.processJsonRpcWithHttp)
	s.httpServer.GET(api.BuiltInPathWsJsonRPC, s.processJsonRpcWithWS)
	s.httpServer.GET(api.BuiltInPathRawWS, s.processRawWS)

	if s.opt.EnableInnerService {
		s.httpServer.GET(api.BuiltInPathMetrics, g.WrapH(promhttp.Handler()))

		path, err := filepath.Abs(filepath.Dir(os.Args[0]))
		gobase.True(err == nil)
		dir := http.Dir(path + "/..")
		log.Infof("set static fs to %s", dir)
		s.httpServer.StaticFS(api.BuiltInPathDIR, dir)

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

func (s *Server) RegisterRawWs(handle api.RawWsHandle) error {
	s.rawWsHandle = handle
	return nil
}

func (s *Server) RegisterPath(httpMethod string, path string, handle api.RawHttpHandle) error {
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
		rootPath == api.BuiltInPathWsJsonRPC ||
		rootPath == api.BuiltInPathRawWS ||
		rootPath == api.BuiltInPathDIR {
		return invalidPath
	}

	return nil
}

func (s *Server) Run(ln net.Listener) error {
	return s.httpServer.RunListener(ln)
}

func (s *Server) processRawWS(c *g.Context) {
	gobase.True(s.rawWsHandle != nil)
	srv, err := websocket.NewServer(c, websocket.WithRawHandle(s.rawWsHandle))
	if err != nil {
		c.String(http.StatusBadRequest, "websocket.NewServer: %s", err)
		return
	}
	defer srv.Close()
	srv.Run()
}

func (s *Server) processJsonRpcWithWS(c *g.Context) {
	gobase.True(s.rpcServer != nil)
	srv, err := websocket.NewServer(c, websocket.WithRpcServer(s.rpcServer))
	if err != nil {
		c.String(http.StatusBadRequest, "websocket.NewServer: %s", err)
		return
	}
	defer srv.Close()
	srv.Run()
}

func (s *Server) processJsonRpcWithHttp(c *g.Context) {
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
