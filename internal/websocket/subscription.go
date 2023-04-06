package websocket

import "github.com/BabySid/gorpc/api"

var _ api.Notifier = (*notifier)(nil)

type notifier struct {
	s *Server
}

func (n *notifier) Notify(sub *api.SubscriptionNotice) {
	_ = n.s.WriteJson(sub)
}
