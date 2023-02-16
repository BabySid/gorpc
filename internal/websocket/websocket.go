package websocket

import (
	"github.com/gorilla/websocket"
	"sync"
)

const (
	wsReadBuffer       = 4096
	wsWriteBuffer      = 4096
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
