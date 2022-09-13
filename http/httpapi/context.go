package httpapi

import (
	"fmt"
	"github.com/BabySid/gobase"
	"github.com/BabySid/gorpc/monitor"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type APIContext struct {
	Name     string
	RecvTime time.Time
	ID       interface{}

	ctx *gin.Context
}

func DefaultAPIContext(name string) *APIContext {
	traceID := fmt.Sprintf("%s_%s", name, uuid.New().String())
	ctx := &APIContext{"", time.Now(), traceID, nil}
	return ctx
}

func NewAPIContext(name string, id interface{}, reqSize int, c *gin.Context) *APIContext {
	monitor.ProcessingRequests.WithLabelValues(monitor.GetCluster(), name).Inc()
	monitor.RealTimeRequestBodySize.WithLabelValues(monitor.GetCluster(), name).Set(float64(reqSize))
	ctx := &APIContext{name, time.Now(), id, c}
	ctx.ToLog("NewAPIContext: reqSize[%d] clientIP[%s]", reqSize, ctx.ClientIP())
	return ctx
}

func (ctx *APIContext) EndRequest(code int) {
	ctx.ToLog("EndRequest %d", code)

	monitor.ProcessingRequests.WithLabelValues(monitor.GetCluster(), ctx.Name).Dec()
	monitor.TotalRequests.WithLabelValues(monitor.GetCluster(), ctx.Name, fmt.Sprintf("%d", code)).Inc()
	monitor.RequestLatency.WithLabelValues(monitor.GetCluster(), ctx.Name).Observe(float64(time.Since(ctx.RecvTime).Milliseconds()))
	monitor.RealTimeRequestLatency.WithLabelValues(monitor.GetCluster(), ctx.Name).Set(float64(time.Since(ctx.RecvTime).Milliseconds()))
}

func (ctx *APIContext) ToLog(format string, v ...interface{}) {
	log.Infof("%s ID[%v] CostSince[%v]",
		fmt.Sprintf(format, v...), ctx.ID, time.Since(ctx.RecvTime))
}

func (ctx *APIContext) ClientIP() string {
	gobase.True(ctx.ctx != nil)
	return ctx.ctx.ClientIP()
}

func (ctx *APIContext) ReplyJson(obj interface{}) {
	gobase.True(ctx.ctx != nil)
	ctx.ctx.JSON(http.StatusOK, obj)
	return
}

func (ctx *APIContext) DefaultQuery(key, defValue string) string {
	gobase.True(ctx.ctx != nil)
	return ctx.ctx.DefaultQuery(key, defValue)
}
