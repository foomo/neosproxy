package logging

import (
	"bytes"
	"log"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger
var Stacks = false
var initializeOnce sync.Once

func init() {
	initializeOnce.Do(initLog)
}

func initLog() {
	Logger = logrus.New()
	Logger.Formatter = &logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	}
	Logger.Level = logrus.DebugLevel
	// Logger.Formatter = &logrus.TextFormatter{
	// 	ForceColors: true,
	// }
	log.SetOutput(&logWriter{})
}

type logWriter struct{}

func (lw *logWriter) Write(p []byte) (n int, err error) {
	getFuncLogEntry(4).Info(string(p))
	return len(p), nil
}

func getFuncLogEntry(skip int) Entry {
	f := "unknown"
	file := "unknown"
	line := -1
	pc, file, line, ok := runtime.Caller(skip)
	if ok {
		f = runtime.FuncForPC(pc).Name()
	}
	return &entry{
		stack: Stacks,
		l: Logger.WithFields(logrus.Fields{
			"func": f,
			"file": file,
			"line": line,
		}),
	}
}

func LogPanic(r interface{}) {
	err := GetErrorFromRecovery(r)
	buf := new(bytes.Buffer)
	if err := pprof.Lookup("goroutine").WriteTo(buf, 1); err != nil {
		buf.String()
	}

	logrus.WithFields(logrus.Fields{
		"stack":    string(debug.Stack()),
		"routines": buf.String(),
	}).Error(err.Error())
}
