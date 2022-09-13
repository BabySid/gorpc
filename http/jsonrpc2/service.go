package jsonrpc2

import (
	"fmt"
	"reflect"
)

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

func (s *service) call(mtype *methodType, argv, replyv reflect.Value) {
	function := mtype.method.Func
	// Invoke the method, providing a new value for the reply.
	returnValues := function.Call([]reflect.Value{s.receiver, argv, replyv})
	// The return value for the method is an error.
	errInter := returnValues[0].Interface()
	errmsg := ""
	if errInter != nil {
		errmsg = errInter.(error).Error()
	}
	fmt.Println(errmsg, replyv.Interface())
}
