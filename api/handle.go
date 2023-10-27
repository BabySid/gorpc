package api

// RpcHandle is a raw interface for creating api based http
type RpcHandle func(ctx Context, params interface{}) *JsonRpcResponse

type Notifier interface {
	Notify(sub *SubscriptionNotice) error
	Err() chan error
}

const (
	NotifierKey = "_notifierKey_"
)
