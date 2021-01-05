package logger

import (
	"encoding/json"
	"fmt"

	"github.com/lonng/nano/log"

	"github.com/lonng/nano/x/hook"
	"github.com/sirupsen/logrus"
)

const (
	Spk = "spk"
	Opt = "opt"
	Sys = "sys"
	Run = "run"
)

// Logger is a wrapper for logrus.Logger
type Logger struct {
	*logrus.Logger
	fields map[string]string
}

// NewLogger return a new logger initialized by name config
func NewLogger() *Logger {
	return &Logger{logrus.New(), nil}
}

func (l *Logger) ReadLevel(name string, c map[string]interface{}) error {
	v, ok := c["level"]
	if !ok {
		return nil
	}
	level, ok := v.(string)
	if !ok {
		return fmt.Errorf("Logger %s level %v is not string, ignoring it", name, v)
	}
	l.SetLevel(parseLevel(level))
	return nil
}

// Hook creates all hooks for logger, and attaches hooks to it.
func (l *Logger) ReadHooks(name string, c map[string]interface{}, processors map[string]*hook.Processor) error {
	v, ok := c["hooks"]
	if !ok {
		return nil
	}
	hooks, ok := v.(map[string]interface{})
	if !ok {
		return fmt.Errorf("Logger %s hooks %v is not string, ignoring it", name, v)
	}

	for typ, v := range hooks {
		s, err := json.Marshal(v)
		if err != nil {
			return err
		}

		h, err := hook.New(name, typ, processors[typ], s)
		if err != nil {
			log.Warnf("Logger %s Hook %s is not found, ignoring it", name, typ)
			continue
		}
		l.Hooks.Add(h)
	}
	return nil
}

// SetReplaceFields set replace map for fields
func (l *Logger) SetReplaceFields(m map[string]string) {
	l.fields = m
}

func (l *Logger) replaceFields(fields logrus.Fields) logrus.Fields {
	if l.fields == nil {
		return fields
	}
	newFields := logrus.Fields{}
	for key, v := range fields {
		if newKey, ok := l.fields[key]; ok {
			newFields[newKey] = v
		} else {
			newFields[key] = v
		}
	}
	return newFields
}

// LogFields use fields to store log data
func (l *Logger) LogFields(level logrus.Level,
	fields logrus.Fields, args ...interface{}) {
	l.WithFields(l.replaceFields(fields)).Log(level, args...)
}
