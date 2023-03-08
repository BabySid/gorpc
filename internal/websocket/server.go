package websocket

import (
	"fmt"
	"github.com/BabySid/gorpc/internal/jsonrpc"
	ws "github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	conn *ws.Conn

	wg        sync.WaitGroup
	readErr   chan error
	readOp    chan readOp
	closeCh   chan struct{}
	pingReset chan struct{}
}

func NewServer(w http.ResponseWriter, r *http.Request) (*Server, error) {
	s := Server{}

	conn, err := upGrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	s.conn = conn
	s.readErr = make(chan error)
	s.closeCh = make(chan struct{})
	s.pingReset = make(chan struct{})
	s.conn.SetReadLimit(wsMessageSizeLimit)
	s.conn.SetPongHandler(func(_ string) error {
		s.conn.SetReadDeadline(time.Time{})
		return nil
	})

	s.wg.Add(1)
	go s.pingLoop()
	return &s, nil
}

func (s *Server) WriteJson(v interface{}) error {
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

	// dispatch
	// readMessage (client request)
	// sendMessage (server response or subscription)
	// if close, cancel all the subscription
	for {
		select {
		case <-s.closeCh:
			return
		case <-s.readErr:
			return
		case op := <-s.readOp:
			// handle(op)
			fmt.Println(op)
			return
			//case write
			//case read
		}
	}
}

func (s *Server) read() {
	for {
		msgs, batch, err := s.readBatch()
		if err != nil {
			s.readErr <- err
			return
		}
		s.readOp <- readOp{
			msgs:  msgs,
			batch: batch,
		}
	}
}

func (s *Server) readBatch() ([]*jsonrpc.Message, bool, error) {
	_, data, err := s.conn.ReadMessage()
	if err != nil {
		return nil, false, err
	}

	msgs, batch, err := jsonrpc.ParseBatchMessage(data)
	if err != nil {
		return nil, false, err
	}

	return msgs, batch, nil
}

type readOp struct {
	msgs  []*jsonrpc.Message
	batch bool
}
