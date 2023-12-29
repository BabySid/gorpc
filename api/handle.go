package api

// RawHandle is a raw interface for creating api based http
type RawHandle func(ctx RawContext, body []byte)

type Notifier interface {
	Notify(sub *SubscriptionNotice) error
	Err() chan error
}

const (
	NotifierKey = "_notifierKey_"
)
