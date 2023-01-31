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
	g "google.golang.org/grpc"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type ServerOption struct {
	Addr        string
	ClusterName string
	// logs
	Rotator  *log.Rotator
	LogLevel string

	HttpOpt httpcfg.ServerOption

	BeforeRun func() error
}

type Server struct {
	option ServerOption

	httpServer *http.Server
	grpcServer *grpc.Server
	mux        cmux.CMux
	pidFile    string
}

func NewServer(opt ServerOption) *Server {
	log.InitLog(opt.LogLevel, opt.Rotator)

	s := &Server{
		option:     opt,
		httpServer: http.NewServer(opt.HttpOpt),
		grpcServer: grpc.NewServer(),
	}
	return s
}

func (s *Server) RegisterJsonRPC(name string, receiver interface{}) error {
	return s.httpServer.RegisterJsonRPC(name, receiver)
}

func (s *Server) RegisterPath(httpMethod string, path string, handle httpapi.RpcHandle) error {
	return s.httpServer.RegisterPath(httpMethod, path, handle)
}

func (s *Server) RegisterGrpc(desc *g.ServiceDesc, impl interface{}) error {
	return s.grpcServer.RegisterGRPC(desc, impl)
}

func (s *Server) Run() error {
	monitor.InitMonitor(s.option.ClusterName)

	if s.option.BeforeRun != nil {
		if err := s.option.BeforeRun(); err != nil {
			l.Warnf("run handle failed. err: %v", err)
			return err
		}
	}

	ln, err := net.Listen("tcp", s.option.Addr)
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

	if err = s.mux.Serve(); err != nil {
		// https://github.com/soheilhy/cmux/issues/39
		if strings.Contains(err.Error(), "use of closed network connection") {
			return nil
		}

		return err
	}

	return nil
}

func (s *Server) Stop() error {
	s.mux.Close()
	_ = os.Remove(s.pidFile)
	return nil
}
