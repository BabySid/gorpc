package websocket

import "github.com/BabySid/gorpc/api"

var _ api.JsonRpcNotifier = (*rpcNotifier)(nil)

type rpcNotifier struct {
	s  *Server
	id string
}

func (n *rpcNotifier) ID() string {
	return n.id
}

func (n *rpcNotifier) Err() chan error {
	return n.s.serverErr
}

func (n *rpcNotifier) Notify(sub *api.SubscriptionNotice) error {
	return n.s.writeJson(sub)
}

var _ api.RawWSNotifier = (*rawNotifier)(nil)

type rawNotifier struct {
	s  *Server
	id string
}

func (r *rawNotifier) Write(msg api.WSMessage) error {
	return r.s.writeRaw(int(msg.Type), msg.Data)
}

func (r *rawNotifier) ID() string {
	return r.id
}

func (r *rawNotifier) Err() chan error {
	return r.s.serverErr
}
