package gin

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

const (
	logDateTimeLayout = "2006-01-02 15:04:05.000"
)

var logFormatter = func(param gin.LogFormatterParams) string {
	return fmt.Sprintf("INFO %s - %s %s %s %s %d %s %d [%s]\n",
		param.TimeStamp.Format(logDateTimeLayout),
		param.ClientIP,
		param.Method,
		param.Path,
		param.Request.Proto,
		param.StatusCode,
		param.Latency,
		param.BodySize,
		param.Request.UserAgent(),
	)
}
