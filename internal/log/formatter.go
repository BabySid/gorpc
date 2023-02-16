package log

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type formatter struct{}

func newFormatter() *formatter {
	return &formatter{}
}

const (
	logDateTimeLayout = "2006-01-02 15:04:05.000"
)

func (s *formatter) Format(entry *log.Entry) ([]byte, error) {
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
