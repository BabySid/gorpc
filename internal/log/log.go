package log

import (
	"log/slog"

	"github.com/BabySid/gobase/log"
)

var logHandler log.Logger

func InitLog(log log.Logger) {
	logHandler = log
}

func Trace(msg string, attrs ...slog.Attr) {
	logHandler.Trace(msg, attrs...)
}

func Debug(msg string, attrs ...slog.Attr) {
	logHandler.Debug(msg, attrs...)
}

func Info(msg string, attrs ...slog.Attr) {
	logHandler.Info(msg, attrs...)
}

func Warn(msg string, attrs ...slog.Attr) {
	logHandler.Warn(msg, attrs...)
}

func Error(msg string, attrs ...slog.Attr) {
	logHandler.Error(msg, attrs...)
}

func OutLogger() *slog.Logger {
	return logHandler.OutLogger()
}

func ErrLogger() *slog.Logger {
	return logHandler.ErrLogger()
}
