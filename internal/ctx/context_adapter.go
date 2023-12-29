package ctx

import (
	"fmt"
	"github.com/BabySid/gorpc/api"
	"github.com/BabySid/gorpc/metrics"
	log "github.com/sirupsen/logrus"
	"time"
)

var _ api.Context = (*ContextAdapter)(nil)

type ContextAdapter struct {
	Name    string
	RevTime time.Time
	ID      interface{}

	KV map[string]any
}

func (ctx *ContextAdapter) ClientIP() string {
	//TODO implement me
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
	ctx.Log("EndRequest %d", code)

	metrics.ProcessingRequests.WithLabelValues(metrics.GetCluster(), ctx.Name).Dec()
	metrics.TotalRequests.WithLabelValues(metrics.GetCluster(), ctx.Name, fmt.Sprintf("%d", code)).Inc()
	metrics.RequestLatency.WithLabelValues(metrics.GetCluster(), ctx.Name).Observe(float64(time.Since(ctx.RevTime).Milliseconds()))
	metrics.RealTimeRequestLatency.WithLabelValues(metrics.GetCluster(), ctx.Name).Set(float64(time.Since(ctx.RevTime).Milliseconds()))
}

func (ctx *ContextAdapter) Log(format string, v ...interface{}) {
	log.Infof("%s Name[%s] CtxID[%v] CostSince[%v]",
		fmt.Sprintf(format, v...), ctx.Name, ctx.CtxID(), time.Since(ctx.RevTime))
}
