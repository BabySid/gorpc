package gorpc

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/BabySid/gorpc/api"
	"github.com/BabySid/gorpc/internal/grpc"
	"github.com/BabySid/gorpc/internal/http"
	"github.com/BabySid/gorpc/internal/log"
	"github.com/BabySid/gorpc/metrics"
	"github.com/soheilhy/cmux"
	g "google.golang.org/grpc"
)

type Server struct {
	option api.ServerOption

	hSvr *http.Server
	gSvr *grpc.Server
	mux  cmux.CMux

	pidFile string
	pid     int
	netFile string

	stopOnce sync.Once
}

func NewServer(opt api.ServerOption) *Server {
	log.InitLog(opt.Logger)

	s := &Server{
		option: opt,
		hSvr:   http.NewServer(opt),
		gSvr:   grpc.NewServer(),
	}
	return s
}

func (s *Server) RegisterJsonRPC(name string, receiver interface{}) error {
	return s.hSvr.RegisterJsonRPC(name, receiver)
}

func (s *Server) RegisterPath(httpMethod string, path string, handle api.RawHttpHandle) error {
	return s.hSvr.RegisterPath(httpMethod, path, handle)
}

func (s *Server) RegisterRawWs(handle api.RawWsHandle) error {
	return s.hSvr.RegisterRawWs(handle)
}

func (s *Server) RegisterGrpc(desc *g.ServiceDesc, impl interface{}) error {
	return s.gSvr.RegisterGRPC(desc, impl)
}

func (s *Server) Run() error {
	metrics.InitMonitor(s.option.ClusterName)

	if s.option.BeforeRun != nil {
		if err := s.option.BeforeRun(); err != nil {
			log.DefaultLog.Warn("run handle failed", slog.Any("err", err))
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
		_ = s.gSvr.Run(grpcL)
	}()

	go func() {
		httpL := s.mux.Match(cmux.HTTP1Fast())
		_ = s.hSvr.Run(httpL)
	}()

	s.pidFile = fmt.Sprintf("%s.pid", filepath.Base(os.Args[0]))
	s.pid = os.Getpid()
	_ = os.WriteFile(s.pidFile, []byte(strconv.Itoa(s.pid)), 0o666)

	s.netFile = fmt.Sprintf("%s.net", filepath.Base(os.Args[0]))
	_ = os.WriteFile(s.netFile, []byte(ln.Addr().String()), 0o666)

	log.DefaultLog.Info("gorpc server begin to run", slog.Any("lnAddr", ln.Addr()), slog.Int("pid", s.pid))

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
	s.stopOnce.Do(func() {
		log.DefaultLog.Info("gorpc server stopped", slog.Int("pid", s.pid))
		if s.pidFile != "" {
			_ = os.Remove(s.pidFile)
		}
		if s.netFile != "" {
			_ = os.Remove(s.netFile)
		}
		if s.mux != nil {
			s.mux.Close()
		}
	})
	return nil
}
