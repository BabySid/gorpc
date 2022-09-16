package jsonrpc2

import (
	"bytes"
	"encoding/json"
	"github.com/BabySid/gorpc/http/httpapi"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"go/token"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
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

func stdParamsDecoder(raw interface{}, params interface{}) error {
	bs, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bs, params)
	return err
}

func pbParamsDecoder(raw interface{}, params interface{}) error {
	bs, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	newReader, err := utilities.IOReaderFactory(bytes.NewReader(bs))
	if err != nil {
		return err
	}

	defaultMarshal := &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			EmitUnpopulated: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}

	err = defaultMarshal.NewDecoder(newReader()).Decode(params)
	if err != nil && err != io.EOF {
		return err
	}

	return nil
}
