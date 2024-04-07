package websocket

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/BabySid/gobase"
	"github.com/BabySid/gorpc/api"
	"github.com/BabySid/gorpc/internal/jsonrpc"
	"github.com/BabySid/gorpc/internal/log"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
)

type Server struct {
	conn *ws.Conn

	wg        sync.WaitGroup
	readErr   chan error
	readOp    chan api.WSMessage
	closeCh   chan struct{}
	pingReset chan struct{}

	lastErr   error
	serverErr chan error

	wMux sync.Mutex

	ctx *gin.Context

	clientIP string
	option   wsOption
}

type wsOption struct {
	rpcServer   *jsonrpc.Server
	rpcNotifier *rpcNotifier

	rawHandle   api.RawWsHandle
	rawNotifier *rawNotifier
}

type WsOption func(opt *wsOption)

func WithRpcServer(s *jsonrpc.Server) WsOption {
	return func(opt *wsOption) {
		gobase.True(opt.rawHandle == nil)
		opt.rpcServer = s
	}
}

func WithRawHandle(handle api.RawWsHandle) WsOption {
	return func(opt *wsOption) {
		gobase.True(opt.rpcServer == nil)
		opt.rawHandle = handle
	}
}

func NewServer(ctx *gin.Context, opts ...WsOption) (*Server, error) {
	gobase.True(len(opts) > 0)

	conn, err := upGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return nil, err
	}

	s := Server{}
	for _, opt := range opts {
		opt(&s.option)
	}

	s.conn = conn
	s.readErr = make(chan error)
	s.readOp = make(chan api.WSMessage)
	s.closeCh = make(chan struct{})
	s.pingReset = make(chan struct{})
	s.serverErr = make(chan error, 1)
	s.conn.SetReadLimit(wsMessageSizeLimit)
	s.conn.SetPongHandler(func(v string) error {
		_ = s.conn.SetReadDeadline(time.Time{})
		return nil
	})

	s.ctx = ctx
	if s.option.rpcServer != nil {
		s.option.rpcNotifier = &rpcNotifier{
			s:  &s,
			id: uuid.New().String(),
		}
	}
	if s.option.rawHandle != nil {
		s.option.rawNotifier = &rawNotifier{
			s:  &s,
			id: uuid.New().String(),
		}
	}

	s.clientIP = ctx.ClientIP()

	log.DefaultLog.Info("connect to websocket", slog.String("clientIP", s.clientIP))

	s.wg.Add(1)
	go s.pingLoop()
	return &s, nil
}

func (s *Server) writeJson(v interface{}) error {
	s.wMux.Lock()
	defer s.wMux.Unlock()

	_ = s.conn.SetWriteDeadline(time.Now().Add(wsWriteTimeout))
	err := s.conn.WriteJSON(v)
	if err == nil {
		s.pingReset <- struct{}{}
	}

	return err
}

func (s *Server) writeRaw(typ int, data []byte) error {
	s.wMux.Lock()
	defer s.wMux.Unlock()

	_ = s.conn.SetWriteDeadline(time.Now().Add(wsWriteTimeout))
	err := s.conn.WriteMessage(typ, data)
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
	s.serverErr <- s.lastErr
	log.DefaultLog.Info("close from websocket", slog.String("clientIP", s.clientIP))
}

func (s *Server) pingLoop() {
	timer := time.NewTimer(wsPingInterval)
	defer s.wg.Done()
	defer timer.Stop()

	for {
		select {
		case <-s.closeCh:
			log.DefaultLog.Debug("recv closeCh in pingLoop", slog.String("clientIP", s.clientIP))
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
			log.DefaultLog.Debug("recv closeCh in Run", slog.String("clientIP", s.clientIP))
			return
		case err := <-s.readErr:
			log.DefaultLog.Debug("recv readErr in Run", slog.String("clientIP", s.clientIP), slog.Any("err", err))
			s.lastErr = err
			return
		case msg := <-s.readOp:
			if s.option.rawHandle != nil {
				_ = s.handleRaw(msg)
			} else {
				_ = s.handleJsonRpc(msg)
			}

		}
	}
}

func (s *Server) read() {
	for {
		typ, data, err := s.conn.ReadMessage()
		if err != nil {
			log.DefaultLog.Debug("recv readErr in read", slog.String("clientIP", s.clientIP), slog.Any("err", err))
			s.readErr <- err
			return
		}
		s.readOp <- api.WSMessage{Type: api.WSMessageType(typ), Data: data}
	}
}

func (s *Server) handleRaw(msg api.WSMessage) error {
	context := newWSContext("RawWs", uuid.New().String(), len(msg.Data), s)
	defer func() {
		context.EndRequest(api.Success)
	}()

	context.WithValue(api.RawWSNotifierKey, s.option.rawNotifier)
	return s.option.rawHandle(context, msg)
}

func (s *Server) handleJsonRpc(msg api.WSMessage) error {
	context := newWSContext("jsonRpc2", uuid.New().String(), len(msg.Data), s)
	defer func() {
		context.EndRequest(api.Success)
	}()

	context.WithValue(api.JsonRpcNotifierKey, s.option.rpcNotifier)

	resp := s.option.rpcServer.Call(context, msg.Data)
	return s.writeJson(resp)
}
