package logging

import (
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/pkg/errors"
)

const (
	stackFieldKey = "stack"
)

type Entry interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatalln(args ...interface{})
	Panic(args ...interface{})

	SetLogger(logger *logrus.Logger)
	SetField(name string, value interface{})
	HasField(name string) bool

	WithError(err error) Entry
	WithFirstError(err ...error) Entry
	WithFields(logrus.Fields) Entry
	WithField(name string, field interface{}) Entry
	WithStack(skip int) Entry
	WithDuration(start time.Time) Entry
}

type entry struct {
	stack bool
	l     *logrus.Entry
}

func (e *entry) SetLogger(logger *logrus.Logger) {
	e.l.Logger = logger
}

func (e *entry) withStack() *entry {
	if e.stack {
		return e.WithField(stackFieldKey, getStack(2)).(*entry)
	}
	return e
}

func getStack(skip int) string {
	skip++

	const size = 64 << 10
	buf := make([]byte, size)
	buf = buf[:runtime.Stack(buf, false)]
	var lines []string
	stackLines := strings.Split(string(buf), "\n")
	if len(stackLines) > 0 {
		stackLines = stackLines[1:]
	}
	for i, line := range stackLines {
		if i/2 > skip {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

func (e *entry) WithStack(skip int) Entry {
	return e.WithField(stackFieldKey, getStack(skip))
}

func (e *entry) Debug(args ...interface{}) {
	e.withStack().l.Debug(args...)
}

func (e *entry) Debugf(format string, args ...interface{}) {
	e.withStack().l.Debugf(format, args...)
}

func (e *entry) Warnf(format string, args ...interface{}) {
	e.withStack().l.Warnf(format, args...)
}

func (e *entry) Fatalln(args ...interface{}) {
	e.withStack().l.Fatalln(args...)
}

func (e *entry) Info(args ...interface{}) {
	e.withStack().l.Info(args...)
}

func (e *entry) Infof(format string, args ...interface{}) {
	e.withStack().l.Infof(format, args...)
}

func (e *entry) Infoln(args ...interface{}) {
	e.withStack().l.Infoln(args...)
}

func (e *entry) Warn(args ...interface{}) {
	e.withStack().l.Warn(args...)
}

func (e *entry) Error(args ...interface{}) {
	e.withStack().l.Error(args...)
}
func (e *entry) Errorf(format string, args ...interface{}) {
	e.withStack().l.Errorf(format, args...)
}

func (e *entry) Fatal(args ...interface{}) {
	e.withStack().l.Fatal(args...)
}

func (e *entry) Panic(args ...interface{}) {
	e.withStack().l.Panic(args...)
}

func (e *entry) Fatalf(format string, args ...interface{}) {
	e.withStack().l.Fatalf(format, args...)
}

func (e *entry) WithError(err error) Entry {
	return &entry{
		l: e.l.WithError(err),
	}
}

func (e *entry) WithFirstError(errors ...error) Entry {
	var err error
	for _, e := range errors {
		if e != nil {
			err = e
			break
		}
	}

	return &entry{
		l: e.l.WithError(err),
	}
}

func (e *entry) WithWrappedError(err error, message string) Entry {
	return &entry{
		l: e.l.WithError(errors.Wrap(err, message)),
	}
}

func (e *entry) WithFields(fields logrus.Fields) Entry {
	return &entry{
		l: e.l.WithFields(fields),
	}
}
func (e *entry) WithField(name string, field interface{}) Entry {
	return &entry{
		l: e.l.WithField(name, field),
	}
}

func (e *entry) WithDuration(start time.Time) Entry {
	return &entry{
		l: e.l.WithField(FieldDuration, time.Since(start).Seconds()),
	}
}

func (e *entry) HasField(name string) bool {
	_, has := e.l.Data[name]
	return has
}

func (e *entry) SetField(name string, value interface{}) {
	e.l.Data[name] = value
}
