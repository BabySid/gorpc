package ctx

import (
	"fmt"
	"github.com/BabySid/gobase"
	"github.com/BabySid/gorpc/api"
	"github.com/BabySid/gorpc/metrics"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"time"
)

var _ api.Context = (*APIContext)(nil)

type APIContext struct {
	name    string
	revTime time.Time
	id      interface{}

	ctx *gin.Context
	kv  map[string]any
}

func (ctx *APIContext) WithValue(key string, value any) {
	ctx.kv[key] = value
}

func (ctx *APIContext) Value(key string) (any, bool) {
	v, ok := ctx.kv[key]
	return v, ok
}

func (ctx *APIContext) ID() interface{} {
	return ctx.id
}

func DefaultAPIContext(name string) *APIContext {
	traceID := fmt.Sprintf("%s_%s", name, uuid.New().String())
	ctx := &APIContext{name: "", revTime: time.Now(), id: traceID, ctx: nil, kv: make(map[string]any)}
	return ctx
}

func NewAPIContext(name string, id interface{}, reqSize int, c *gin.Context) *APIContext {
	metrics.ProcessingRequests.WithLabelValues(metrics.GetCluster(), name).Inc()
	metrics.RealTimeRequestBodySize.WithLabelValues(metrics.GetCluster(), name).Set(float64(reqSize))
	ctx := &APIContext{name: name, revTime: time.Now(), id: id, ctx: c, kv: make(map[string]any)}
	ctx.Log("NewAPIContext: reqSize[%d] clientIP[%s]", reqSize, ctx.ClientIP())
	return ctx
}

func (ctx *APIContext) EndRequest(code int) {
	ctx.Log("EndRequest %d", code)

	metrics.ProcessingRequests.WithLabelValues(metrics.GetCluster(), ctx.name).Dec()
	metrics.TotalRequests.WithLabelValues(metrics.GetCluster(), ctx.name, fmt.Sprintf("%d", code)).Inc()
	metrics.RequestLatency.WithLabelValues(metrics.GetCluster(), ctx.name).Observe(float64(time.Since(ctx.revTime).Milliseconds()))
	metrics.RealTimeRequestLatency.WithLabelValues(metrics.GetCluster(), ctx.name).Set(float64(time.Since(ctx.revTime).Milliseconds()))
}

func (ctx *APIContext) Log(format string, v ...interface{}) {
	log.Infof("%s Name[%s] ID[%v] CostSince[%v]",
		fmt.Sprintf(format, v...), ctx.name, ctx.id, time.Since(ctx.revTime))
}

func (ctx *APIContext) ClientIP() string {
	gobase.True(ctx.ctx != nil)
	return ctx.ctx.ClientIP()
}