package jsonrpc

import (
	"encoding/json"
	"github.com/BabySid/gorpc/api"
)

var null = json.RawMessage("null")

// Message A value of this type can a JSON-RPC request, notification, successful response or
// error response. Which one it is depends on the fields.
type Message struct {
	Version string `json:"jsonrpc,omitempty"`
	// The type of the ID field must be json.Message because the interface{} type cannot perform type comparisons correctly during the process.
	// For example, in respWait(sync.Map), the request is of type int, but it becomes float64 when in response.
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	// If this is an instance of Subscription, the Params is json.Marshal(api.SubscriptionResult)
	Params json.RawMessage   `json:"params,omitempty"`
	Error  *api.JsonRpcError `json:"error,omitempty"`
	Result json.RawMessage   `json:"result,omitempty"`
}

func (msg *Message) IsNotification() bool {
	return msg.ID == nil && msg.Method != ""
}

func (msg *Message) IsCall() bool {
	return msg.HasValidID() && msg.Method != ""
}

func (msg *Message) IsResponse() bool {
	return msg.HasValidID() && msg.Method == "" && msg.Params == nil && (msg.Result != nil || msg.Error != nil)
}

func (msg *Message) HasValidID() bool {
	return len(msg.ID) > 0 && msg.ID[0] != '{' && msg.ID[0] != '['
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
