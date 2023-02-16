package jsonrpc

import (
	"encoding/json"
	"github.com/BabySid/gorpc/api"
)

var null = json.RawMessage("null")

type SubscriptionResult struct {
	ID     interface{}     `json:"subscription"`
	Result json.RawMessage `json:"result,omitempty"`
}

// Message A value of this type can a JSON-RPC request, notification, successful response or
// error response. Which one it is depends on the fields.
type Message struct {
	Version string            `json:"jsonrpc,omitempty"`
	ID      interface{}       `json:"id,omitempty"`
	Method  string            `json:"method,omitempty"`
	Params  json.RawMessage   `json:"params,omitempty"`
	Error   *api.JsonRpcError `json:"error,omitempty"`
	Result  json.RawMessage   `json:"result,omitempty"`
}

func (msg *Message) isNotification() bool {
	return msg.ID == nil && msg.Method != ""
}

func (msg *Message) isCall() bool {
	return msg.hasValidID() && msg.Method != ""
}

func (msg *Message) isResponse() bool {
	return msg.hasValidID() && msg.Method == "" && msg.Params == nil && (msg.Result != nil || msg.Error != nil)
}

func (msg *Message) hasValidID() bool {
	return msg.ID != nil
}

//func (msg *jsonrpcMessage) isSubscribe() bool {
//	return strings.HasSuffix(msg.Method, subscribeMethodSuffix)
//}
//
//func (msg *jsonrpcMessage) isUnsubscribe() bool {
//	return strings.HasSuffix(msg.Method, unsubscribeMethodSuffix)
//}
//
//func (msg *jsonrpcMessage) namespace() string {
//	elem := strings.SplitN(msg.Method, serviceMethodSeparator, 2)
//	return elem[0]
//}

func (msg *Message) String() string {
	b, _ := json.Marshal(msg)
	return string(b)
}