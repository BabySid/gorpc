package gin

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	*gin.Engine
}

func NewServer() *Server {
	gin.SetMode(gin.ReleaseMode)

	s := Server{}
	s.Engine = gin.New()
	s.Engine.Use(gin.LoggerWithFormatter(logFormatter))
	s.Engine.Use(gin.Recovery())

	gin.DefaultWriter = log.StandardLogger().Out
	return &s
}
