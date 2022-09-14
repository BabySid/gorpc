package jsonrpc2

import (
	"encoding/json"
	"github.com/BabySid/gorpc/http/httpapi"
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

func parseRequestBody(b []byte) (interface{}, error) {
	var err error
	var jsonData interface{}
	err = json.Unmarshal(b, &jsonData)
	if err != nil {
		return nil, err
	}
	return jsonData, err
}

func parseRequestMap(jsonMap map[string]interface{}) (*httpapi.JsonRpcRequest, *httpapi.JsonRpcError) {
	var request httpapi.JsonRpcRequest
	request.Id = jsonMap["id"]

	if v, ok := jsonMap["jsonrpc"]; ok {
		if ver, ok := v.(string); ok {
			request.Version = ver
		}
	}
	if m, ok := jsonMap["method"]; ok {
		if method, ok := m.(string); ok {
			request.Method = method
		}
	}
	request.Params = jsonMap["params"]

	if request.Version != httpapi.Version {
		return nil, httpapi.NewJsonRpcError(httpapi.InvalidRequest, httpapi.SysCodeMap[httpapi.InvalidRequest], "invalid version")
	}

	if request.Id == nil {
		return nil, httpapi.NewJsonRpcError(httpapi.InvalidRequest, httpapi.SysCodeMap[httpapi.InvalidRequest], "id must set")
	}

	return &request, nil
}

func parseRequestParams(raw interface{}, params interface{}) error {
	bs, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bs, params)
	return err
}
