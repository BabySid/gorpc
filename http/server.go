package http

import (
	"github.com/BabySid/gobase"
	"github.com/BabySid/gorpc/http/httpapi"
	"github.com/BabySid/gorpc/http/httpcfg"
	"github.com/BabySid/gorpc/http/jsonrpc2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

type Server struct {
	httpServer *gin.Engine

	rpcServer *jsonrpc2.Server

	// handles
	getHandles  map[string]httpapi.RpcHandle
	postHandles map[string]httpapi.RpcHandle
}

func NewServer(option httpcfg.ServerOption) *Server {
	gin.SetMode(gin.ReleaseMode)

	s := &Server{
		httpServer:  gin.Default(),
		rpcServer:   jsonrpc2.NewServer(option),
		getHandles:  make(map[string]httpapi.RpcHandle),
		postHandles: make(map[string]httpapi.RpcHandle),
	}

	s.httpServer.POST("/", s.processPostRequest)
	return s
}

func (s *Server) RegisterJsonRPC(name string, receiver interface{}) error {
	return s.rpcServer.RegisterName(name, receiver)
}

func (s *Server) RegisterPath(httpMethod string, path string, handle httpapi.RpcHandle) error {
	switch httpMethod {
	case http.MethodGet:
		s.httpServer.GET(path, s.internalHandle)
		s.getHandles[path] = handle
	case http.MethodPost:
		s.httpServer.POST(path, s.internalHandle)
		s.postHandles[path] = handle
	default:
		gobase.AssertHere()
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
	s.httpServer.StaticFS("_dir_", dir)

	// monitor
	s.httpServer.GET("metrics", gin.WrapH(promhttp.Handler()))
	return s.httpServer.RunListener(ln)
}

func (s *Server) internalHandle(c *gin.Context) {
	httpMethod := c.Request.Method

	switch httpMethod {
	case http.MethodGet:
		s.processGetRequest(c)
	case http.MethodPost:
		s.processPostRequest(c)
	default:
		gobase.AssertHere()
	}
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
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		resp := httpapi.NewErrorJsonRpcResponse(nil, httpapi.InternalError, httpapi.SysCodeMap[httpapi.InternalError], err.Error())
		c.JSON(http.StatusOK, resp)
		return
	}

	path := c.Request.URL.Path
	if path == "" || path == "/" {
		s.processJsonRpc(c, body)
		return
	}

	var req httpapi.JsonRpcRequest
	err = httpapi.DecodeJson(body, &req)
	if err != nil {
		resp := httpapi.NewErrorJsonRpcResponse(nil, httpapi.ParseError, httpapi.SysCodeMap[httpapi.ParseError], err.Error())
		c.JSON(http.StatusOK, resp)
		return
	}

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

func (s *Server) processJsonRpc(c *gin.Context, body []byte) {
	ctx := httpapi.NewAPIContext("jsonRpc2", uuid.New().String(), len(body), c)
	defer func() {
		ctx.EndRequest(httpapi.Success)
	}()

	resp := s.rpcServer.Call(ctx, body)
	c.JSON(http.StatusOK, resp)
}
