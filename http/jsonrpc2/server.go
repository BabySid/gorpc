package jsonrpc2

import (
	"errors"
	"fmt"
	"github.com/BabySid/gorpc/http/base"
	log "github.com/sirupsen/logrus"
	"go/token"
	"reflect"
	"strings"
	"sync"
)

// Precompute the reflect type for error. Can't use error directly
// because Typeof takes an empty interface value. This is annoying.
var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

// Server represents an RPC Server.
type Server struct {
	serviceMap sync.Map // map[string]*service
}

// NewServer returns a new Server.
func NewServer() *Server {
	return &Server{}
}

// DefaultServer is the default instance of *Server.
var DefaultServer = NewServer()

// Register publishes the receiver's methods in the DefaultServer.
func Register(rcvr interface{}) error { return DefaultServer.Register(rcvr) }

// RegisterName is like Register but uses the provided name for the type
// instead of the receiver's concrete type.
func RegisterName(name string, rcvr interface{}) error {
	return DefaultServer.RegisterName(name, rcvr)
}

func (server *Server) Register(rcvr interface{}) error {
	return server.register(rcvr, "", false)
}

// RegisterName is like Register but uses the provided name for the type
// instead of the receiver's concrete type.
func (server *Server) RegisterName(name string, rcvr interface{}) error {
	return server.register(rcvr, name, true)
}

func (server *Server) register(rcvr interface{}, name string, useName bool) error {
	s := new(service)
	s.typ = reflect.TypeOf(rcvr)
	s.receiver = reflect.ValueOf(rcvr)
	sname := name
	if !useName {
		sname = reflect.Indirect(s.receiver).Type().Name()
	}
	if sname == "" {
		return errors.New("rpc.Register: no service name for type " + s.typ.String())
	}
	if !useName && !token.IsExported(sname) {
		return errors.New("rpc.Register: type " + sname + " is not exported")
	}
	s.name = sname

	// Install the methods
	s.method = suitableMethods(s.typ)

	if len(s.method) == 0 {
		str := ""

		// To help the user, see if a pointer receiver would work.
		method := suitableMethods(reflect.PtrTo(s.typ))
		if len(method) != 0 {
			str = "rpc.Register: type " + sname + " has no exported methods of suitable type (hint: pass a pointer to value of that type)"
		} else {
			str = "rpc.Register: type " + sname + " has no exported methods of suitable type"
		}
		return errors.New(str)
	}

	if _, dup := server.serviceMap.LoadOrStore(sname, s); dup {
		return errors.New("rpc: service already defined: " + sname)
	}
	return nil
}

// suitableMethods returns suitable Rpc methods of typ. It will log
// errors if logErr is true.
func suitableMethods(typ reflect.Type) map[string]*methodType {
	methods := make(map[string]*methodType)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		// Method must be exported.
		if !method.IsExported() {
			continue
		}
		// Method needs three ins: receiver, *args, *reply.
		if mtype.NumIn() != 3 {
			log.Warnf("rpc.Register: method %q has %d input parameters; needs exactly three\n", mname, mtype.NumIn())
			continue
		}
		// First arg need not be a pointer.
		argType := mtype.In(1)
		if !isExportedOrBuiltinType(argType) {
			log.Warnf("rpc.Register: argument type of method %q is not exported: %q\n", mname, argType)
			continue
		}
		// Second arg must be a pointer.
		replyType := mtype.In(2)
		if replyType.Kind() != reflect.Ptr {
			log.Warnf("rpc.Register: reply type of method %q is not a pointer: %q\n", mname, replyType)
			continue
		}
		// Reply type must be exported.
		if !isExportedOrBuiltinType(replyType) {
			log.Warnf("rpc.Register: reply type of method %q is not exported: %q\n", mname, replyType)
			continue
		}
		// Method needs one out.
		if mtype.NumOut() != 1 {
			log.Warnf("rpc.Register: method %q has %d output parameters; needs exactly one\n", mname, mtype.NumOut())
			continue
		}
		// The return type of the method must be error.
		if returnType := mtype.Out(0); returnType != typeOfError {
			log.Warnf("rpc.Register: return type of method %q is %q, must be error\n", mname, returnType)
			continue
		}
		methods[mname] = &methodType{method: method, ArgType: argType, ReplyType: replyType}
	}
	return methods
}

func (s *Server) Call(data []byte) error {
	var req base.JsonRpcRequest
	err := base.DecodeJson(data, &req)
	if err != nil {
		return err
	}
	dot := strings.LastIndex(req.Method, ".")
	if dot < 0 {
		return errors.New("rpc: service/method request ill-formed: " + req.Method)
	}
	serviceName := req.Method[:dot]
	methodName := req.Method[dot+1:]

	// Look up the request.
	svci, ok := s.serviceMap.Load(serviceName)
	if !ok {
		return errors.New("rpc: can't find service " + req.Method)
	}
	svc := svci.(*service)
	mtype := svc.method[methodName]
	if mtype == nil {
		return errors.New("rpc: can't find method " + req.Method)
	}

	argIsValue := false // if true, need to indirect before calling.
	var argv, replyv reflect.Value
	if mtype.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(mtype.ArgType.Elem())
	} else {
		argv = reflect.New(mtype.ArgType)
		argIsValue = true
	}
	if argIsValue {
		argv = argv.Elem()
	}

	// argv guaranteed to be a pointer now.
	if err = base.DecodeJson(*req.Params, argv.Interface()); err != nil {
		return err
	}

	replyv = reflect.New(mtype.ReplyType.Elem())

	switch mtype.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(mtype.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(mtype.ReplyType.Elem(), 0, 0))
	}

	svc.call(mtype, argv, replyv)
	fmt.Println(replyv.String(), replyv.Elem().String())
	return nil
}
