package websocket

import (
	"log/slog"
	"time"

	"github.com/BabySid/gobase"
	"github.com/BabySid/gorpc/api"
	"github.com/BabySid/gorpc/internal/ctx"
	"github.com/BabySid/gorpc/internal/log"
	"github.com/BabySid/gorpc/metrics"
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
			Logger:  nil,
		},
	}

	wsCtx.Logger = log.OutLogger().With("name", wsCtx.Name, "ctxID", wsCtx.ID, "clientIP", wsCtx.ClientIP())
	wsCtx.Logger.Info("NewWSContext", slog.Int("reqSize", reqSize))
	return wsCtx
}
