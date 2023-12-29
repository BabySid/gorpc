package api

type Context interface {
	CtxID() interface{}
	Log(format string, v ...interface{})
	ClientIP() string
	EndRequest(code int)
	WithValue(key string, value any)
	Value(key string) (any, bool)
}

type RawContext interface {
	Context
	Param(key string) string
	Query(key string) string
	WriteData(code int, contType string, data []byte) error
}
