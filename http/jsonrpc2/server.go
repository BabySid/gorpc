package jsonrpc2

import (
	"errors"
	"github.com/BabySid/gobase"
	"github.com/BabySid/gorpc/http/httpapi"
	"github.com/BabySid/gorpc/http/httpcfg"
	log "github.com/sirupsen/logrus"
	"go/token"
	"reflect"
	"strings"
	"sync"
)

// Precompute the reflect type for error. Can't use error directly
// because Typeof takes an empty interface value. This is annoying.
var typeOfError = reflect.TypeOf((*httpapi.JsonRpcError)(nil))

// Server represents an RPC Server.
type Server struct {
	opt        httpcfg.ServerOption
	serviceMap sync.Map // map[string]*service
}

// NewServer returns a new Server.
func NewServer(opt httpcfg.ServerOption) *Server {
	return &Server{opt: opt}
}

// DefaultServer is the default instance of *Server.
var DefaultServer = NewServer(httpcfg.DefaultOption)

// Register publishes the receiver's methods in the DefaultServer.
func Register(receiver interface{}) error { return DefaultServer.Register(receiver) }

// RegisterName is like Register but uses the provided name for the type
// instead of the receiver's concrete type.
func RegisterName(name string, receiver interface{}) error {
	return DefaultServer.RegisterName(name, receiver)
}

func (server *Server) Register(receiver interface{}) error {
	return server.register(receiver, "", false)
}

// RegisterName is like Register but uses the provided name for the type
// instead of the receiver's concrete type.
func (server *Server) RegisterName(name string, receiver interface{}) error {
	return server.register(receiver, name, true)
}

func (server *Server) register(receiver interface{}, name string, useName bool) error {
	s := new(service)
	s.typ = reflect.TypeOf(receiver)
	s.receiver = reflect.ValueOf(receiver)
	serverName := name
	if !useName {
		serverName = reflect.Indirect(s.receiver).Type().Name()
	}
	if serverName == "" {
		return errors.New("rpc.Register: no service name for type " + s.typ.String())
	}
	if !useName && !token.IsExported(serverName) {
		return errors.New("rpc.Register: type " + serverName + " is not exported")
	}
	s.name = serverName

	// Install the methods
	s.method = suitableMethods(s.typ)

	if len(s.method) == 0 {
		str := ""

		// To help the user, see if a pointer receiver would work.
		method := suitableMethods(reflect.PtrTo(s.typ))
		if len(method) != 0 {
			str = "rpc.Register: type " + serverName + " has no exported methods of suitable type (hint: pass a pointer to value of that type)"
		} else {
			str = "rpc.Register: type " + serverName + " has no exported methods of suitable type"
		}
		return errors.New(str)
	}

	// todo register multi method
	if _, dup := server.serviceMap.LoadOrStore(serverName, s); dup {
		return errors.New("rpc: service already defined: " + serverName)
	}
	return nil
}

// suitableMethods returns suitable Rpc methods of typ
func suitableMethods(typ reflect.Type) map[string]*methodType {
	methods := make(map[string]*methodType)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mType := method.Type
		mName := method.Name
		// Method must be exported.
		if !method.IsExported() {
			continue
		}
		// Method needs three ins: receiver, ctx, *args.
		if mType.NumIn() != 3 {
			log.Warnf("rpc.Register: method %q has %d input parameters; needs exactly three\n", mName, mType.NumIn())
			continue
		}

		ctxType := mType.In(1)
		if ctxType.Kind() != reflect.Ptr {
			log.Warnf("rpc.Register: ctx type of method %q is not a pointer: %q\n", mName, ctxType)
			continue
		}
		if ctxType.String() != "*httpapi.APIContext" {
			log.Warnf("rpc.Register: ctx type of method %q is not *httpapi.APIContext: %q\n", mName, ctxType)
			continue
		}

		// First arg need not be a pointer.
		argType := mType.In(2)
		if !isExportedOrBuiltinType(argType) {
			log.Warnf("rpc.Register: argument type of method %q is not exported: %q\n", mName, argType)
			continue
		}

		// Method needs two out.
		if mType.NumOut() != 2 {
			log.Warnf("rpc.Register: method %q has %d input parameters; needs exactly two\n", mName, mType.NumOut())
			continue
		}

		// reply must be a pointer.
		replyType := mType.Out(0)
		if replyType.Kind() != reflect.Ptr {
			log.Warnf("rpc.Register: reply type of method %q is not a pointer: %q\n", mName, replyType)
			continue
		}
		// Reply type must be exported.
		if !isExportedOrBuiltinType(replyType) {
			log.Warnf("rpc.Register: reply type of method %q is not exported: %q\n", mName, replyType)
			continue
		}

		// The return type of the method must be error.
		if returnType := mType.Out(1); returnType != typeOfError {
			log.Warnf("rpc.Register: return type of method %q is %q, must be %q\n", mName, returnType, typeOfError.String())
			continue
		}
		methods[mName] = &methodType{method: method, ArgType: argType, ReplyType: replyType}
	}
	return methods
}

func (server *Server) Call(ctx *httpapi.APIContext, data []byte) interface{} {
	reqData, err := parseRequestBody(data)
	if err != nil {
		return httpapi.NewErrorJsonRpcResponseWithError(nil,
			httpapi.NewJsonRpcError(httpapi.ParseError, httpapi.SysCodeMap[httpapi.ParseError], err.Error()))
	}

	if reflect.ValueOf(reqData).Kind() == reflect.Map {
		input, ok := reqData.(map[string]interface{})
		if !ok {
			return httpapi.NewErrorJsonRpcResponseWithError(nil,
				httpapi.NewJsonRpcError(httpapi.InvalidRequest, httpapi.SysCodeMap[httpapi.InvalidRequest], nil))
		}
		res := server.processRequest(ctx, input)
		return res
	} else if reflect.ValueOf(reqData).Kind() == reflect.Slice {
		resArr := make([]interface{}, 0)
		for _, req := range reqData.([]interface{}) {
			input, ok := req.(map[string]interface{})
			if !ok {
				return httpapi.NewErrorJsonRpcResponseWithError(nil,
					httpapi.NewJsonRpcError(httpapi.InvalidRequest, httpapi.SysCodeMap[httpapi.InvalidRequest], nil))
			}
			res := server.processRequest(ctx, input)
			resArr = append(resArr, res)
		}
		if len(resArr) == 0 {
			return httpapi.NewErrorJsonRpcResponseWithError(nil,
				httpapi.NewJsonRpcError(httpapi.InvalidRequest, httpapi.SysCodeMap[httpapi.InvalidRequest], "empty request"))
		}
		return resArr
	}

	return httpapi.NewErrorJsonRpcResponseWithError(nil,
		httpapi.NewJsonRpcError(httpapi.InvalidRequest, httpapi.SysCodeMap[httpapi.InvalidRequest], "request must array or object"))
}

func (server *Server) processRequest(ctx *httpapi.APIContext, reqMap map[string]interface{}) *httpapi.JsonRpcResponse {
	req, rpcErr := parseRequestMap(reqMap)
	if rpcErr != nil {
		return httpapi.NewErrorJsonRpcResponseWithError(req.Version, rpcErr)
	}

	dot := strings.Index(req.Method, ".")
	if dot < 0 {
		return httpapi.NewErrorJsonRpcResponseWithError(req.Id, httpapi.NewJsonRpcError(httpapi.InvalidRequest,
			httpapi.SysCodeMap[httpapi.InvalidRequest],
			"rpc: service/method request ill-formed: "+req.Method))
	}
	serviceName := req.Method[:dot]
	methodName := req.Method[dot+1:]

	// Look up the request.
	srv, ok := server.serviceMap.Load(serviceName)
	if !ok {
		return httpapi.NewErrorJsonRpcResponseWithError(req.Id, httpapi.NewJsonRpcError(httpapi.MethodNotFound,
			httpapi.SysCodeMap[httpapi.MethodNotFound],
			"rpc: can't find service: "+req.Method))
	}

	svc := srv.(*service)
	mType := svc.method[methodName]
	if mType == nil {
		return httpapi.NewErrorJsonRpcResponseWithError(req.Id, httpapi.NewJsonRpcError(httpapi.MethodNotFound,
			httpapi.SysCodeMap[httpapi.MethodNotFound],
			"rpc: can't find method: "+req.Method))
	}

	argIsValue := false // if true, need to indirect before calling.
	var argv reflect.Value
	if mType.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(mType.ArgType.Elem())
	} else {
		argv = reflect.New(mType.ArgType)
		argIsValue = true
	}
	if argIsValue {
		argv = argv.Elem()
	}

	// argv guaranteed to be a pointer now.
	if err := server.opt.PDecoder(req.Params, argv.Interface()); err != nil {
		return httpapi.NewErrorJsonRpcResponseWithError(req.Id, httpapi.NewJsonRpcError(httpapi.InvalidParams,
			httpapi.SysCodeMap[httpapi.InvalidParams],
			err.Error()))
	}

	//replyValue := reflect.New(mType.ReplyType.Elem())
	//
	//switch mType.ReplyType.Elem().Kind() {
	//case reflect.Map:
	//	replyValue.Elem().Set(reflect.MakeMap(mType.ReplyType.Elem()))
	//case reflect.Slice:
	//	replyValue.Elem().Set(reflect.MakeSlice(mType.ReplyType.Elem(), 0, 0))
	//}

	replyValue, err := svc.call(mType, reflect.ValueOf(ctx), argv)
	apiErr := err.(*httpapi.JsonRpcError)
	if apiErr != nil {
		gobase.TrueF(httpapi.CheckCode(apiErr.Code), "%d conflict with sys error code", apiErr.Code)
		return httpapi.NewErrorJsonRpcResponseWithError(req.Id, apiErr)
	}

	return httpapi.NewSuccessJsonRpcResponse(req.Id, replyValue)
}
