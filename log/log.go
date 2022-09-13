package log

import (
	"github.com/gin-gonic/gin"
	logRotator "github.com/lestrrat-go/file-rotatelogs"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Rotator struct {
	LogMaxAge int
	LogPath   string
}

func InitLog(level string, rotator *Rotator) {
	// because rotate cannot work on windows, so we need a flag to disable it
	if rotator != nil {
		maxAge := rotator.LogMaxAge
		if maxAge <= 0 {
			maxAge = 24 * 7
		}

		path := rotator.LogPath
		if path == "" {
			path = "./"
		}
		path += filepath.Base(os.Args[0]) + ".log"

		writer, _ := logRotator.New(
			path+".%Y%m%d%H",
			logRotator.WithLinkName(path),
			logRotator.WithMaxAge(time.Duration(maxAge)*time.Hour),
			logRotator.WithRotationTime(time.Hour),
		)

		log.SetOutput(writer)
	}

	log.SetReportCaller(true)
	log.SetLevel(getLogLevel(level))
	log.SetFormatter(NewFormatter())

	gin.DefaultWriter = log.StandardLogger().Out
}

func getLogLevel(lv string) log.Level {
	level := strings.ToLower(lv)
	switch level {
	case "trace":
		return log.TraceLevel
	case "debug":
		return log.DebugLevel
	default:
		return log.InfoLevel
	}
}
