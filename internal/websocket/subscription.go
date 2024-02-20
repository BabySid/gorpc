package websocket

import "github.com/BabySid/gorpc/api"

var _ api.Notifier = (*notifier)(nil)

type notifier struct {
	s  *Server
	id string
}

func (n *notifier) ID() string {
	return n.id
}

func (n *notifier) Err() chan error {
	return n.s.notifyErr
}

func (n *notifier) Notify(sub *api.SubscriptionNotice) error {
	return n.s.WriteJson(sub)
}
