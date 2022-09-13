package http

import (
	"github.com/BabySid/gobase"
	"github.com/BabySid/gorpc/http/base"
	"github.com/gin-gonic/gin"
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

	// handles
	getHandles  map[string]base.RpcHandle
	postHandles map[string]base.RpcHandle
}

func NewServer() *Server {
	return &Server{
		httpServer:  gin.Default(),
		getHandles:  make(map[string]base.RpcHandle),
		postHandles: make(map[string]base.RpcHandle),
	}
}

func (s *Server) RegisterJsonRPC(name string, receiver interface{}) error {

	return nil
}

func (s *Server) RegisterPath(httpMethod string, path string, handle base.RpcHandle) error {
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
	s.httpServer.StaticFS("openapi", dir)

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

	id := interface{}(nil)
	if v, ok := c.GetQuery("id"); ok {
		id = v
	}
	handle, ok := s.getHandles[path]
	if !ok {
		resp := base.NewErrorJsonRpcResponse(id, MethodNotFound, sysCodeMap[MethodNotFound], path)
		c.JSON(http.StatusOK, resp)
		return
	}

	code := Success
	ctx := NewAPIContext(path, id, 0, c)
	defer func() {
		gobase.TrueF(checkCode(code), "%d conflict with sys error code", code)
		ctx.EndRequest(code)
	}()

	resp, rpcErr := handle(ctx, nil)
	if rpcErr != nil {
		resp := base.NewErrorJsonRpcResponseWithError(id, rpcErr)
		code = rpcErr.Code
		c.JSON(http.StatusOK, resp)
		return
	}
	c.JSON(http.StatusOK, base.NewSuccessJsonRpcResponse(id, resp))
}

func (s *Server) processPostRequest(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		resp := base.NewErrorJsonRpcResponse(nil, InternalError, sysCodeMap[InternalError], err)
		c.JSON(http.StatusOK, resp)
	}

	var req base.JsonRpcRequest
	err = base.DecodeJson(body, &req)
	if err != nil {
		resp := base.NewErrorJsonRpcResponse(nil, ParseError, sysCodeMap[ParseError], err)
		c.JSON(http.StatusOK, resp)
	}

	path := c.Request.URL.Path

	handle, ok := s.postHandles[path]
	if !ok {
		resp := base.NewErrorJsonRpcResponse(req.Id, MethodNotFound, sysCodeMap[MethodNotFound], path)
		c.JSON(http.StatusOK, resp)
		return
	}

	code := Success
	ctx := NewAPIContext(path, req.Id, len(body), c)
	defer func() {
		gobase.TrueF(checkCode(code), "%d conflict with sys error code", code)
		ctx.EndRequest(code)
	}()

	resp, rpcErr := handle(ctx, req.Params)
	if rpcErr != nil {
		resp := base.NewErrorJsonRpcResponseWithError(req.Id, rpcErr)
		code = rpcErr.Code
		c.JSON(http.StatusOK, resp)
		return
	}
	c.JSON(http.StatusOK, base.NewSuccessJsonRpcResponse(req.Id, resp))
}
