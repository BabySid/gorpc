package ctx

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/BabySid/gorpc/api"
	"github.com/BabySid/gorpc/metrics"
)

var _ api.Context = (*ContextAdapter)(nil)

type ContextAdapter struct {
	Name    string
	RevTime time.Time
	ID      interface{}

	KV map[string]any

	Logger *slog.Logger
}

func (ctx *ContextAdapter) ClientIP() string {
	// TODO implement me
	panic("implement me")
}

func (ctx *ContextAdapter) WithValue(key string, value any) {
	ctx.KV[key] = value
}

func (ctx *ContextAdapter) Value(key string) (any, bool) {
	v, ok := ctx.KV[key]
	return v, ok
}

func (ctx *ContextAdapter) CtxID() interface{} {
	return ctx.ID
}

func (ctx *ContextAdapter) EndRequest(code int) {
	ctx.Logger.Info("EndRequest", slog.Int("code", code), slog.Int("cost", int(time.Since(ctx.RevTime))))

	metrics.ProcessingRequests.WithLabelValues(metrics.GetCluster(), ctx.Name).Dec()
	metrics.TotalRequests.WithLabelValues(metrics.GetCluster(), ctx.Name, fmt.Sprintf("%d", code)).Inc()
	metrics.RequestLatency.WithLabelValues(metrics.GetCluster(), ctx.Name).Observe(float64(time.Since(ctx.RevTime).Milliseconds()))
	metrics.RealTimeRequestLatency.WithLabelValues(metrics.GetCluster(), ctx.Name).Set(float64(time.Since(ctx.RevTime).Milliseconds()))
}
