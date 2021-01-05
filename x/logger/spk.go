package logger

import (
	"bytes"
	"io/ioutil"
	"reflect"

	"github.com/sirupsen/logrus"
)

// NewSpkLogger creates a logger hooked by tcplog.
// which produces logs that are stored in aws's S3 production
// It is analysised by aws EMR and result is stored in mysql.
func NewSpkLogger(c map[string]interface{}) *Logger {
	logger := NewLogger()

	logger.SetFormatter(&SpkFormatter{})
	logger.SetOutput(ioutil.Discard)

	if err := logger.ReadLevel(Sys, c); err != nil {
		panic(err)
	}
	if err := logger.ReadHooks(Spk, c, nil); err != nil {
		panic(err)
	}

	return logger
}

type SpkFormatter struct{}

func (f *SpkFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	bytes, ok := entry.Data["Bytes"]
	if ok {
		return bytes.([]byte), nil
	}

	header := entry.Data["Header"]
	event := entry.Data["Event"]

	bytes, err := f.formatBytes(header, event)
	if err != nil {
		return nil, err
	}
	entry.Data["Bytes"] = bytes

	return bytes.([]byte), nil
}

func (f *SpkFormatter) formatBytes(header interface{}, event interface{}) ([]byte, error) {
	var b bytes.Buffer

	value := reflect.ValueOf(header).Elem()
	for i := 0; i < value.NumField(); i++ {
		b.WriteString(MustString(value.Field(i).Interface()))
		b.WriteByte('|')
	}

	b.WriteByte('^')

	value = reflect.ValueOf(event).Elem()
	for i := 0; i < value.NumField(); i++ {
		b.WriteByte('|')
		b.WriteString(MustString(value.Field(i).Interface()))
	}
	b.WriteByte('\n')

	return b.Bytes(), nil
}
