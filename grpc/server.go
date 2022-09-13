package grpc

import (
	"google.golang.org/grpc"
	"net"
)

type Server struct {
	gServer *grpc.Server
}

func NewServer() *Server {
	var opts []grpc.ServerOption
	return &Server{gServer: grpc.NewServer(opts...)}
}

func (s *Server) RegisterGRPC(desc *grpc.ServiceDesc, impl interface{}) error {
	s.gServer.RegisterService(desc, impl)
	return nil
}

func (s *Server) Run(ln net.Listener) error {
	return s.gServer.Serve(ln)
}
