package log

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Formatter struct{}

func NewFormatter() *Formatter {
	return &Formatter{}
}

const (
	logDateTimeLayout = "\"2006-01-02 15:04:05.000\""
)

func (s *Formatter) Format(entry *log.Entry) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(strings.ToUpper(entry.Level.String()) + " ")
	buf.WriteString(time.Now().Local().Format(logDateTimeLayout) + " ")

	var file string
	var len int
	if entry.Caller != nil {
		file = filepath.Base(entry.Caller.File)
		len = entry.Caller.Line
	}
	buf.WriteString("[" + file + ":" + strconv.Itoa(len) + "]" + " ")

	buf.WriteString("{")
	flag := false
	for k, v := range entry.Data {
		if flag {
			buf.WriteString("; ")
		}
		buf.WriteString(k + ":")
		buf.WriteString(fmt.Sprintf("%v", v))
		flag = true
	}
	buf.WriteString("}" + " ")

	buf.WriteString(entry.Message)

	buf.WriteString("\n")

	return buf.Bytes(), nil
}

var GinLogFormatter = func(param gin.LogFormatterParams) string {
	return fmt.Sprintf("INFO %s - %s %s %s %s %d %s %d [%s]\n",
		param.TimeStamp.Format(logDateTimeLayout),
		param.ClientIP,
		param.Method,
		param.Path,
		param.Request.Proto,
		param.StatusCode,
		param.Latency,
		param.BodySize,
		param.Request.UserAgent(),
	)
}
