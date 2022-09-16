package gorpc

import (
	"github.com/BabySid/gorpc/grpc"
	"github.com/BabySid/gorpc/http"
	"github.com/BabySid/gorpc/http/httpapi"
	"github.com/BabySid/gorpc/http/httpcfg"
	"github.com/BabySid/gorpc/log"
	"github.com/BabySid/gorpc/monitor"
	l "github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
	"net"
)

type ServerOption struct {
	Addr        string
	ClusterName string
	// logs
	Rotator  *log.Rotator
	LogLevel string
}

type Server struct {
	httpServer *http.Server
	grpcServer *grpc.Server
}

func NewServer(opt httpcfg.ServerOption) *Server {
	s := &Server{
		httpServer: http.NewServer(opt),
		grpcServer: grpc.NewServer(),
	}
	return s
}

func (s *Server) RegisterJsonRPC(name string, receiver interface{}) error {
	s.httpServer.RegisterJsonRPC(name, receiver)
	return nil
}

func (s *Server) RegisterPath(httpMethod string, path string, handle httpapi.RpcHandle) error {
	return s.httpServer.RegisterPath(httpMethod, path, handle)
}

func (s *Server) Run(option ServerOption) error {
	log.InitLog(option.LogLevel, option.Rotator)
	monitor.InitMonitor(option.ClusterName)

	ln, err := net.Listen("tcp", option.Addr)
	if err != nil {
		return err
	}

	m := cmux.New(ln)

	go func() {
		grpcL := m.MatchWithWriters(
			cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
		_ = s.grpcServer.Run(grpcL)
	}()

	go func() {
		httpL := m.Match(cmux.HTTP1Fast())
		_ = s.httpServer.Run(httpL)
	}()

	l.Infof("gorpc server run on %s", ln.Addr())
	return m.Serve()
}
