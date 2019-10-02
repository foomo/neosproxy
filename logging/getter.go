package logging

import (
	"errors"

	"github.com/sirupsen/logrus"
)

func GetDefaultLogEntry() Entry {
	return GetLogEntry(logrus.Fields{})
}

func GetLogEntry(fields logrus.Fields) Entry {
	return &entry{
		stack: Stacks,
		l:     Logger.WithFields(fields),
	}
}

func GetFuncLogEntry() Entry {
	if Stacks {
		return GetLogEntry(logrus.Fields{})
	}
	return getFuncLogEntry(1)
}

func GetErrorFromRecovery(r interface{}) (err error) {
	switch x := r.(type) {
	case string:
		err = errors.New(x)
	case error:
		err = x
	default:
		err = errors.New("Unknown panic occurred")
	}
	return
}
