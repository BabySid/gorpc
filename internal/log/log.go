package log

import (
	"github.com/BabySid/gobase/log"
)

var DefaultLog log.Logger

func InitLog(log log.Logger) {
	if log != nil {
		DefaultLog = log
		return
	}
	DefaultLog = defaultLogger()
}

func defaultLogger() log.Logger {
	l := log.NewSLogger(log.WithOutFile("./gorpc.log"), log.WithLevel("trace"))
	return l
}
