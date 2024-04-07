package api

type Context interface {
	CtxID() interface{}
	ClientIP() string
	WithValue(key string, value any)
	Value(key string) (any, bool)
}

type RawHttpContext interface {
	Context
	Param(key string) string
	Query(key string) string
	WriteData(code int, contType string, data []byte) error
}
