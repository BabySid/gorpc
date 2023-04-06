package api

import (
	"io"
)

type ClientType int

const (
	HttpClient = iota
	GrpcClient
	WsClient
)

// BatchElem is an element in a batch request.
type BatchElem struct {
	Method string
	Args   interface{}
	// The result is unmarshalled into this field. Result must be set to a
	// non-nil pointer value of the desired type, otherwise the response will be
	// discarded.
	Result interface{}
	// Error is set if the server returns an error for this request, or if
	// unmarshalling into Result fails. It is not set for I/O errors.
	Error error
}

type Client interface {
	GetType() ClientType
	CallJsonRpc(result interface{}, method string, args interface{}) error
	BatchCallJsonRpc(b []BatchElem) error
	RawCallHttp(method string, path string, body interface{}) (int, io.ReadCloser, error)
	Close() error

	// The WriteByWs ending with "WS" are only intended for WebSocket clients
	WriteByWs(req interface{}) error
	ErrFromWS() chan error

	// The UnderlyingHandle is only intended for grpc clients now
	UnderlyingHandle() interface{}
}

var _ Client = (*ClientAdapter)(nil)

type ClientAdapter struct{}

func (c ClientAdapter) ErrFromWS() chan error {
	//TODO implement me
	panic("implement me")
}

func (c ClientAdapter) WriteByWs(req interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (c ClientAdapter) GetType() ClientType {
	//TODO implement me
	panic("implement me")
}

func (c ClientAdapter) CallJsonRpc(result interface{}, method string, args interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (c ClientAdapter) BatchCallJsonRpc(b []BatchElem) error {
	//TODO implement me
	panic("implement me")
}

func (c ClientAdapter) RawCallHttp(method string, path string, body interface{}) (int, io.ReadCloser, error) {
	//TODO implement me
	panic("implement me")
}

func (c ClientAdapter) UnderlyingHandle() interface{} {
	//TODO implement me
	panic("implement me")
}

func (c ClientAdapter) Close() error {
	//TODO implement me
	panic("implement me")
}
