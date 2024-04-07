package gin

import (
	"github.com/gin-gonic/gin"
)

type Server struct {
	*gin.Engine
}

func NewServer() *Server {
	gin.SetMode(gin.ReleaseMode)

	s := Server{}
	s.Engine = gin.New()
	s.Engine.Use(ginLogger())
	s.Engine.Use(gin.Recovery())

	// gin.DefaultWriter = io.MultiWriter(log.StandardLogger().Out)
	// gin.DefaultErrorWriter = io.MultiWriter(log.StandardLogger().Out)
	return &s
}
