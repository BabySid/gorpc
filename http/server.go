package http

import (
	"errors"
	"fmt"
	"github.com/BabySid/gobase"
	"github.com/BabySid/gorpc/http/httpapi"
	"github.com/BabySid/gorpc/http/httpcfg"
	"github.com/BabySid/gorpc/http/jsonrpc2"
	l "github.com/BabySid/gorpc/log"
	"github.com/gin-gonic/gin"
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
	httpServer *gin.Engine

	rpcServer *jsonrpc2.Server

	// handles
	getHandles  map[string]httpapi.RpcHandle
	postHandles map[string]httpapi.RpcHandle
}

const (
	BuiltInJsonRPC = "_jsonrpc_"
	BuiltInDIR     = "_dir_"
	BuiltInMetrics = "_metrics_"
)

func NewServer(option httpcfg.ServerOption) *Server {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.LoggerWithFormatter(l.GinLogFormatter))
	router.Use(gin.Recovery())

	s := &Server{
		httpServer:  router,
		rpcServer:   jsonrpc2.NewServer(option),
		getHandles:  make(map[string]httpapi.RpcHandle),
		postHandles: make(map[string]httpapi.RpcHandle),
	}

	return s
}

func (s *Server) setUpBuiltInService() {
	s.httpServer.POST(BuiltInJsonRPC, s.processJsonRpc)
	s.httpServer.GET(BuiltInMetrics, gin.WrapH(promhttp.Handler()))

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
`, appName, BuiltInDIR, appName, BuiltInMetrics, appName)

	s.httpServer.GET("/", func(ctx *gin.Context) {
		ctx.Header("Content-Type", "text/html; charset=utf-8")

		ctx.String(http.StatusOK, indexHtml)
	})
}

func (s *Server) RegisterJsonRPC(name string, receiver interface{}) error {
	return s.rpcServer.RegisterName(name, receiver)
}

func (s *Server) RegisterPath(httpMethod string, path string, handle httpapi.RpcHandle) error {
	if err := s.checkPath(path); err != nil {
		return err
	}
	switch httpMethod {
	case http.MethodGet:
		s.httpServer.GET(path, s.processGetRequest)
		s.getHandles[path] = handle
	case http.MethodPost:
		s.httpServer.POST(path, s.processPostRequest)
		s.postHandles[path] = handle
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

	if rootPath == BuiltInMetrics || rootPath == BuiltInJsonRPC || rootPath == BuiltInDIR {
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
	s.httpServer.StaticFS(BuiltInDIR, dir)

	return s.httpServer.RunListener(ln)
}

func (s *Server) processGetRequest(c *gin.Context) {
	path := c.Request.URL.Path

	id := uuid.New().String()
	if v, ok := c.GetQuery("id"); ok {
		id = v
	}
	handle, ok := s.getHandles[path]
	if !ok {
		resp := httpapi.NewErrorJsonRpcResponse(id, httpapi.MethodNotFound, httpapi.SysCodeMap[httpapi.MethodNotFound], path)
		c.JSON(http.StatusOK, resp)
		return
	}

	code := httpapi.Success
	ctx := httpapi.NewAPIContext(path, id, 0, c)
	defer func() {
		ctx.EndRequest(code)
	}()

	resp := handle(ctx, nil)
	if resp.Error != nil {
		code = resp.Error.Code
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) processPostRequest(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		resp := httpapi.NewErrorJsonRpcResponse(nil, httpapi.InternalError, httpapi.SysCodeMap[httpapi.InternalError], err.Error())
		c.JSON(http.StatusOK, resp)
		return
	}

	var req httpapi.JsonRpcRequest
	err = httpapi.DecodeJson(body, &req)
	if err != nil {
		resp := httpapi.NewErrorJsonRpcResponse(nil, httpapi.ParseError, httpapi.SysCodeMap[httpapi.ParseError], err.Error())
		c.JSON(http.StatusOK, resp)
		return
	}

	path := c.Request.URL.Path
	handle, ok := s.postHandles[path]
	if !ok {
		resp := httpapi.NewErrorJsonRpcResponse(req.Id, httpapi.MethodNotFound, httpapi.SysCodeMap[httpapi.MethodNotFound], path)
		c.JSON(http.StatusOK, resp)
		return
	}

	code := httpapi.Success
	ctx := httpapi.NewAPIContext(path, req.Id, len(body), c)
	defer func() {
		ctx.EndRequest(code)
	}()

	resp := handle(ctx, req.Params)
	if resp.Error != nil {
		code = resp.Error.Code
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) processJsonRpc(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		resp := httpapi.NewErrorJsonRpcResponse(nil, httpapi.InternalError, httpapi.SysCodeMap[httpapi.InternalError], err.Error())
		c.JSON(http.StatusOK, resp)
		return
	}

	ctx := httpapi.NewAPIContext("jsonRpc2", uuid.New().String(), len(body), c)
	defer func() {
		ctx.EndRequest(httpapi.Success)
	}()

	resp := s.rpcServer.Call(ctx, body)
	c.JSON(http.StatusOK, resp)
}
