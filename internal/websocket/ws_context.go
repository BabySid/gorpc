package websocket

import (
	"github.com/BabySid/gobase"
	"github.com/BabySid/gorpc/api"
	"github.com/BabySid/gorpc/internal/ctx"
	"github.com/BabySid/gorpc/metrics"
	"time"
)

var _ api.Context = (*Context)(nil)

type Context struct {
	srv *Server
	ctx.ContextAdapter
}

func (ctx *Context) ClientIP() string {
	gobase.True(ctx.srv.ctx != nil)
	return ctx.srv.ctx.ClientIP()
}

func newWSContext(name string, id interface{}, reqSize int, s *Server) *Context {
	metrics.ProcessingRequests.WithLabelValues(metrics.GetCluster(), name).Inc()
	metrics.RealTimeRequestBodySize.WithLabelValues(metrics.GetCluster(), name).Set(float64(reqSize))
	wsCtx := &Context{
		srv: s,
		ContextAdapter: ctx.ContextAdapter{
			Name:    name,
			RevTime: time.Now(),
			ID:      id,
			KV:      make(map[string]any),
		},
	}
	wsCtx.Log("NewWSContext: reqSize[%d] clientIP[%s]", reqSize, wsCtx.ClientIP())
	return wsCtx
}
