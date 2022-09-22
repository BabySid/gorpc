package gorpc

import (
	"fmt"
	"github.com/BabySid/gorpc/grpc"
	"github.com/BabySid/gorpc/http"
	"github.com/BabySid/gorpc/http/httpapi"
	"github.com/BabySid/gorpc/http/httpcfg"
	"github.com/BabySid/gorpc/log"
	"github.com/BabySid/gorpc/monitor"
	l "github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
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
	mux        cmux.CMux
	pidFile    string
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

	s.mux = cmux.New(ln)

	go func() {
		grpcL := s.mux.MatchWithWriters(
			cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
		_ = s.grpcServer.Run(grpcL)
	}()

	go func() {
		httpL := s.mux.Match(cmux.HTTP1Fast())
		_ = s.httpServer.Run(httpL)
	}()

	s.pidFile = fmt.Sprintf("%s.pid", filepath.Base(os.Args[0]))
	_ = ioutil.WriteFile(s.pidFile, []byte(strconv.Itoa(os.Getpid())), 0666)

	l.Infof("gorpc server run on %s", ln.Addr())
	return s.mux.Serve()
}

func (s *Server) Stop() error {
	s.mux.Close()
	_ = os.Remove(s.pidFile)
	return nil
}
