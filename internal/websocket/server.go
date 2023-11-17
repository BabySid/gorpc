package websocket

import (
	"errors"
	"fmt"
	"github.com/BabySid/gorpc/api"
	"github.com/BabySid/gorpc/internal/ctx"
	"github.com/BabySid/gorpc/internal/jsonrpc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Server struct {
	conn *ws.Conn

	wg        sync.WaitGroup
	readErr   chan error
	readOp    chan []byte
	closeCh   chan struct{}
	pingReset chan struct{}

	lastErr   error
	notifyErr chan error

	wMux      sync.Mutex
	rpcServer *jsonrpc.Server

	ctx *gin.Context

	clientIP string
}

func NewServer(rpc *jsonrpc.Server, ctx *gin.Context) (*Server, error) {
	log.Infof("create websocket server: clientIP[%s]", ctx.ClientIP())
	s := Server{}

	conn, err := upGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return nil, err
	}

	s.conn = conn
	s.readErr = make(chan error)
	s.readOp = make(chan []byte)
	s.closeCh = make(chan struct{})
	s.pingReset = make(chan struct{})
	s.notifyErr = make(chan error)
	s.conn.SetReadLimit(wsMessageSizeLimit)
	s.conn.SetPongHandler(func(v string) error {
		_ = s.conn.SetReadDeadline(time.Time{})
		return nil
	})

	s.rpcServer = rpc

	s.ctx = ctx

	s.clientIP = ctx.ClientIP()

	s.wg.Add(1)
	go s.pingLoop()
	return &s, nil
}

func (s *Server) WriteJson(v interface{}) error {
	s.wMux.Lock()
	defer s.wMux.Unlock()

	_ = s.conn.SetWriteDeadline(time.Now().Add(wsWriteTimeout))
	err := s.conn.WriteJSON(v)
	if err == nil {
		s.pingReset <- struct{}{}
	}

	return err
}

func (s *Server) Close() {
	close(s.closeCh)
	_ = s.conn.Close()
	s.wg.Wait()

	if s.lastErr == nil {
		s.lastErr = errors.New(fmt.Sprintf("server close from [%s]", s.clientIP))
	}
	s.notifyErr <- s.lastErr
	log.Infof("close websocket server: clientIP[%s]", s.ctx.ClientIP())
}

func (s *Server) pingLoop() {
	var timer = time.NewTimer(wsPingInterval)
	defer s.wg.Done()
	defer timer.Stop()

	for {
		select {
		case <-s.closeCh:
			return
		case <-s.pingReset:
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(wsPingInterval)
		case <-timer.C:
			_ = s.conn.SetWriteDeadline(time.Now().Add(wsPingWriteTimeout))
			_ = s.conn.WriteMessage(ws.PingMessage, nil)
			_ = s.conn.SetReadDeadline(time.Now().Add(wsPongTimeout))
			timer.Reset(wsPingInterval)
		}
	}
}

func (s *Server) Run() {
	go s.read()

	for {
		select {
		case <-s.closeCh:
			log.Tracef("recv closeCh in Run from [%s]", s.clientIP)
			return
		case err := <-s.readErr:
			s.lastErr = err
			return
		case op := <-s.readOp:
			_ = s.handle(op)
		}
	}
}

func (s *Server) read() {
	for {
		_, data, err := s.conn.ReadMessage()
		if err != nil {
			s.readErr <- err
			return
		}
		s.readOp <- data
	}
}

func (s *Server) handle(data []byte) error {
	context := ctx.NewAPIContext("jsonRpc2", uuid.New().String(), len(data), s.ctx)
	defer func() {
		context.EndRequest(api.Success)
	}()

	context.WithValue(api.NotifierKey, &notifier{s: s})

	resp := s.rpcServer.Call(context, data)
	return s.WriteJson(resp)
}
