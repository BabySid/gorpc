package log

import (
	"github.com/BabySid/gobase/log"
)

var DefaultLog log.Logger

func InitLog(log log.Logger) {
	DefaultLog = log
}
