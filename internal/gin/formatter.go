package gin

import (
	"log/slog"
	"time"

	"github.com/BabySid/gorpc/internal/log"
	"github.com/gin-gonic/gin"
)

func ginLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		log.DefaultLog.Info("[GIN]",
			slog.String("clientIP", c.ClientIP()),
			slog.String("method", c.Request.Method),
			slog.String("proto", c.Request.Proto),
			slog.String("path", c.Request.URL.Path),
			slog.String("query", c.Request.URL.RawQuery),
			slog.Int64("contLen", c.Request.ContentLength),
			slog.String("userAgent", c.Request.UserAgent()),
			slog.Int("statusCode", c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
		)
	}
}
