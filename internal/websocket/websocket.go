package websocket

import (
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

const (
	wsReadBuffer       = 4096
	wsWriteBuffer      = 4096
	wsPingInterval     = 60 * time.Second
	wsPingWriteTimeout = 5 * time.Second
	wsPongTimeout      = 30 * time.Second
	wsMessageSizeLimit = 10 * 1024 * 1024
)

var (
	wsBufferPool = new(sync.Pool)
	upGrader     = websocket.Upgrader{
		ReadBufferSize:  wsReadBuffer,
		WriteBufferSize: wsWriteBuffer,
		WriteBufferPool: wsBufferPool,
	}
)
