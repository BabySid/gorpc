package api

// RawHttpHandle is a raw interface for creating api based http
type RawHttpHandle func(RawHttpContext, []byte)

// RawWsHandle is a raw interface for creating api based ws
type RawWsHandle func(Context, WSMessage) error

type JsonRpcNotifier interface {
	Notify(sub *SubscriptionNotice) error
	ID() string
	Err() chan error
}

const (
	JsonRpcNotifierKey = "_JsonRpcNotifierKey_"
	RawWSNotifierKey   = "_RawWSNotifierKey_"
)

type RawWSNotifier interface {
	Write(WSMessage) error
	ID() string
	Err() chan error
}
