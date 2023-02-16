package api

type Context interface {
	ID() interface{}
	Log(format string, v ...interface{})
	ClientIP() string
	EndRequest(code int)
}
