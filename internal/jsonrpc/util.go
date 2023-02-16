package jsonrpc

import (
	"bytes"
	"encoding/json"
	"github.com/BabySid/gorpc/api"
	"go/token"
	"reflect"
)

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return token.IsExported(t.Name()) || t.PkgPath() == ""
}

func parseBatchMessage(b []byte) ([]*Message, bool, error) {
	var rawMsg json.RawMessage
	if err := json.Unmarshal(b, &rawMsg); err != nil {
		return nil, false, err
	}

	messages, batch := parseMessage(rawMsg)
	for i, msg := range messages {
		if msg == nil {
			messages[i] = new(Message)
		}
	}

	return messages, batch, nil
}

// isBatch returns true when the first non-whitespace characters is '['
func isBatch(raw json.RawMessage) bool {
	for _, c := range raw {
		// skip insignificant whitespace (http://www.ietf.org/rfc/rfc4627.txt)
		if c == 0x20 || c == 0x09 || c == 0x0a || c == 0x0d {
			continue
		}
		return c == '['
	}
	return false
}

func parseMessage(raw json.RawMessage) ([]*Message, bool) {
	if !isBatch(raw) {
		msgs := []*Message{{}}
		_ = json.Unmarshal(raw, &msgs[0])
		return msgs, false
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.Token() // skip '['
	var msgs []*Message
	for dec.More() {
		msgs = append(msgs, new(Message))
		dec.Decode(&msgs[len(msgs)-1])
	}
	return msgs, true
}

//func parseRequestBody(b []byte) (interface{}, error) {
//	var err error
//	var jsonData interface{}
//	err = json.Unmarshal(b, &jsonData)
//	if err != nil {
//		return nil, err
//	}
//	return jsonData, err
//}

func checkMessage(message *Message) *api.JsonRpcError {
	if message.Version != api.Version {
		return api.NewJsonRpcError(api.InvalidRequest, api.SysCodeMap[api.InvalidRequest], "invalid version")
	}

	if message.ID == nil {
		return api.NewJsonRpcError(api.InvalidRequest, api.SysCodeMap[api.InvalidRequest], "id must set")
	}

	return nil
}
