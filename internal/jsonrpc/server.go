package jsonrpc

import (
	"errors"
	"go/token"
	"log/slog"
	"reflect"
	"strings"
	"sync"

	"github.com/BabySid/gorpc/api"
	"github.com/BabySid/gorpc/codec"
	"github.com/BabySid/gorpc/internal/log"
)

// Precompute the reflect type for error. Can't use error directly
// because Typeof takes an empty interface value. This is annoying.
var (
	typeOfRpcError = reflect.TypeOf((*api.JsonRpcError)(nil))
	typeOfAPICtx   = reflect.TypeOf((*api.Context)(nil)).Elem()
)

// Server represents an RPC Server.
type Server struct {
	opt        Option
	serviceMap sync.Map // map[string]*service
}

type Option struct {
	CodeType codec.CodecType
}

// NewServer returns a new Server.
func NewServer(opt Option) *Server {
	return &Server{opt: opt}
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
			log.DefaultLog.Warn("rpc.Register failed. number of input needs exactly three", slog.String("methodName", mName), slog.Int("numIn", mType.NumIn()))
			continue
		}

		ctxType := mType.In(1)
		//if ctxType.Kind() != reflect.Ptr {
		//	log.Warnf("rpc.Register: ctx type of method %q is not a pointer: %q\n", mName, ctxType)
		//	continue
		//}
		if ctxType.String() != typeOfAPICtx.String() {
			log.DefaultLog.Warn("rpc.Register failed. type of ctx in method must be "+typeOfAPICtx.String(), slog.String("methodName", mName), slog.String("typeOfCtx", ctxType.String()))
			continue
		}

		// First arg need not be a pointer.
		argType := mType.In(2)
		if !isExportedOrBuiltinType(argType) {
			log.DefaultLog.Warn("rpc.Register failed. argument type of method must be exported", slog.String("methodName", mName), slog.String("typeOfArg", argType.String()))
			continue
		}

		// Method needs two out.
		if mType.NumOut() != 2 {
			log.DefaultLog.Warn("rpc.Register failed. number of output needs exactly two", slog.String("methodName", mName), slog.Int("numOut", mType.NumIn()))
			continue
		}

		// reply must be a pointer.
		replyType := mType.Out(0)
		if replyType.Kind() != reflect.Ptr {
			log.DefaultLog.Warn("rpc.Register failed. reply type[0] of method must be a pointer", slog.String("methodName", mName), slog.String("typeOfReply[0]", replyType.String()))
			continue
		}
		// Reply type must be exported.
		if !isExportedOrBuiltinType(replyType) {
			log.DefaultLog.Warn("rpc.Register failed. reply type[0] of method must be exported", slog.String("methodName", mName), slog.String("typeOfReply[0]", replyType.String()))
			continue
		}

		// The return type of the method must be error.
		if returnType := mType.Out(1); returnType != typeOfRpcError {
			log.DefaultLog.Warn("rpc.Register failed. reply type[1] of method must be "+typeOfRpcError.String(), slog.String("methodName", mName), slog.String("typeOfReply[1]", returnType.String()))
			continue
		}
		methods[mName] = &methodType{method: method, ArgType: argType, ReplyType: replyType}
	}
	return methods
}

func (server *Server) Call(ctx api.Context, data []byte) interface{} {
	msgs, batch, err := ParseBatchMessage(data)
	if err != nil {
		return api.NewErrorJsonRpcResponseWithError(nil,
			api.NewJsonRpcError(api.ParseError, api.SysCodeMap[api.ParseError], err.Error()))
	}

	// todo handle msg.IsNotification() refer go-ethereum/rpc/handler.go
	if batch {
		if len(msgs) == 0 {
			return api.NewErrorJsonRpcResponseWithError(nil,
				api.NewJsonRpcError(api.InvalidRequest, api.SysCodeMap[api.InvalidRequest], "empty request"))
		}
		resArr := make([]interface{}, 0, len(msgs))
		for _, msg := range msgs {
			res := server.processRequest(ctx, msg)
			resArr = append(resArr, res)
		}
		return resArr
	} else {
		res := server.processRequest(ctx, msgs[0])
		return res
	}
	//reqData, err := parseRequestBody(data)
	//if err != nil {
	//	return api.NewErrorJsonRpcResponseWithError(nil,
	//		api.NewJsonRpcError(api.ParseError, api.SysCodeMap[api.ParseError], err.Error()))
	//}

	//if reflect.ValueOf(reqData).Kind() == reflect.Map {
	//	input, ok := reqData.(map[string]interface{})
	//	if !ok {
	//		return api.NewErrorJsonRpcResponseWithError(nil,
	//			api.NewJsonRpcError(api.InvalidRequest, api.SysCodeMap[api.InvalidRequest], nil))
	//	}
	//	res := server.processRequest(ctx, input)
	//	return res
	//} else if reflect.ValueOf(reqData).Kind() == reflect.Slice {
	//	resArr := make([]interface{}, 0)
	//	for _, req := range reqData.([]interface{}) {
	//		input, ok := req.(map[string]interface{})
	//		if !ok {
	//			return api.NewErrorJsonRpcResponseWithError(nil,
	//				api.NewJsonRpcError(api.InvalidRequest, api.SysCodeMap[api.InvalidRequest], nil))
	//		}
	//		res := server.processRequest(ctx, input)
	//		resArr = append(resArr, res)
	//	}
	//	if len(resArr) == 0 {
	//		return api.NewErrorJsonRpcResponseWithError(nil,
	//			api.NewJsonRpcError(api.InvalidRequest, api.SysCodeMap[api.InvalidRequest], "empty request"))
	//	}
	//	return resArr
	//}
}

func (server *Server) processRequest(ctx api.Context, req *Message) *api.JsonRpcResponse {
	rpcErr := checkMessage(req)
	if rpcErr != nil {
		return api.NewErrorJsonRpcResponseWithError(req.ID, rpcErr)
	}
	log.DefaultLog.Debug("processRequest", slog.String("method", req.Method), slog.String("reqId", string(req.ID)))

	dot := strings.Index(req.Method, ".")
	if dot < 0 {
		return api.NewErrorJsonRpcResponseWithError(req.ID, api.NewJsonRpcError(api.InvalidRequest,
			api.SysCodeMap[api.InvalidRequest],
			"rpc: service/method request ill-formed: "+req.Method))
	}
	serviceName := req.Method[:dot]
	methodName := req.Method[dot+1:]

	// Look up the request.
	srv, ok := server.serviceMap.Load(serviceName)
	if !ok {
		return api.NewErrorJsonRpcResponseWithError(req.ID, api.NewJsonRpcError(api.MethodNotFound,
			api.SysCodeMap[api.MethodNotFound],
			"rpc: can't find service: "+req.Method))
	}

	svc := srv.(*service)
	mType := svc.method[methodName]
	if mType == nil {
		return api.NewErrorJsonRpcResponseWithError(req.ID, api.NewJsonRpcError(api.MethodNotFound,
			api.SysCodeMap[api.MethodNotFound],
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
	if req.Params != nil {
		decoder := codec.GetParamDecoder(server.opt.CodeType)
		if err := decoder(req.Params, argv.Interface()); err != nil {
			return api.NewErrorJsonRpcResponseWithError(req.ID, api.NewJsonRpcError(api.InvalidParams,
				api.SysCodeMap[api.InvalidParams],
				err.Error()))
		}
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
	apiErr := err.(*api.JsonRpcError)
	if apiErr != nil {
		return api.NewErrorJsonRpcResponseWithError(req.ID, apiErr)
	}

	return api.NewSuccessJsonRpcResponse(req.ID, replyValue)
}
