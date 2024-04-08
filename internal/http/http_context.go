package http

import (
	"log/slog"
	"time"

	"github.com/BabySid/gobase"
	"github.com/BabySid/gorpc/api"
	"github.com/BabySid/gorpc/internal/ctx"
	"github.com/BabySid/gorpc/internal/log"
	"github.com/BabySid/gorpc/metrics"
	"github.com/gin-gonic/gin"
)

var _ api.Context = (*Context)(nil)

type Context struct {
	ctx *gin.Context
	ctx.ContextAdapter
}

func (ctx *Context) ClientIP() string {
	gobase.True(ctx.ctx != nil)
	return ctx.ctx.ClientIP()
}

func newHttpContext(name string, id interface{}, reqSize int, c *gin.Context) *Context {
	metrics.ProcessingRequests.WithLabelValues(metrics.GetCluster(), name).Inc()
	metrics.RealTimeRequestBodySize.WithLabelValues(metrics.GetCluster(), name).Set(float64(reqSize))
	httpCtx := &Context{
		ctx: c,
		ContextAdapter: ctx.ContextAdapter{
			Name:    name,
			RevTime: time.Now(),
			ID:      id,
			KV:      make(map[string]any),
			Logger:  nil,
		},
	}

	httpCtx.Logger = log.DefaultLog.WithOut(slog.String("name", httpCtx.Name), slog.Any("ctxID", httpCtx.ID), slog.String("clientIP", httpCtx.ClientIP()))
	httpCtx.Logger.Info("NewHttpContext", slog.Int("reqSize", reqSize))
	return httpCtx
}

var _ api.RawHttpContext = (*RawContext)(nil)

type RawContext struct {
	Context
}

func newRawContext(name string, id interface{}, reqSize int, c *gin.Context) *RawContext {
	metrics.ProcessingRequests.WithLabelValues(metrics.GetCluster(), name).Inc()
	metrics.RealTimeRequestBodySize.WithLabelValues(metrics.GetCluster(), name).Set(float64(reqSize))
	rawCtx := &RawContext{
		Context: Context{
			ctx: c,
			ContextAdapter: ctx.ContextAdapter{
				Name:    name,
				RevTime: time.Now(),
				ID:      id,
				KV:      make(map[string]any),
				Logger:  nil,
			},
		},
	}
	rawCtx.Logger = log.DefaultLog.WithOut(slog.String("name", rawCtx.Name), slog.Any("ctxID", rawCtx.ID), slog.String("clientIP", rawCtx.ClientIP()))
	rawCtx.Logger.Info("NewRawContext", slog.Int("reqSize", reqSize))
	return rawCtx
}

func (r *RawContext) Param(key string) string {
	return r.ctx.Param(key)
}

func (r *RawContext) Query(key string) string {
	return r.ctx.Query(key)
}

func (r *RawContext) WriteData(code int, contType string, data []byte) error {
	r.ctx.Data(code, contType, data)
	return nil
}
