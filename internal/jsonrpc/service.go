package jsonrpc

import "reflect"

type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
	//numCalls   uint
}

type service struct {
	name     string                 // name of service
	receiver reflect.Value          // receiver of methods for the service
	typ      reflect.Type           // type of the receiver
	method   map[string]*methodType // registered methods
}

func (s *service) call(mType *methodType, ctx reflect.Value, argv reflect.Value) (interface{}, interface{}) {
	function := mType.method.Func

	returnValues := function.Call([]reflect.Value{s.receiver, ctx, argv})

	reply := returnValues[0].Interface()
	// The return value for the method is an *httpapi.JsonRpcError.
	//if returnValues[1].Interface()
	err := returnValues[1].Interface()
	return reply, err
}
