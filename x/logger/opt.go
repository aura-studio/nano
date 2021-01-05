package logger

import (
	"fmt"
	"io/ioutil"
	"runtime"

	"github.com/sirupsen/logrus"
)

var (
	fields = logrus.FieldMap{
		logrus.FieldKeyLevel: "_level_",
		logrus.FieldKeyTime:  "_time_",
		logrus.FieldKeyFile:  "_caller_",
		logrus.FieldKeyMsg:   "msg",
	}
)

// NewOptLogger creates a logger hooked by lumberjack.Logger,
// which produces logs that are stored in aliyun's sls production.
// It is only used in sls and visited by web browser.
func NewOptLogger(c map[string]interface{}) *Logger {
	logger := NewLogger()
	logger.SetReportCaller(true)
	logger.SetFormatter(&CustomOptFormatter{logrus.JSONFormatter{
		TimestampFormat:   "2006/01/02 15:04:05.0000000", // the "time" field configuratiom
		DisableTimestamp:  false,
		DisableHTMLEscape: true,
		FieldMap:          fields,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return "", fmt.Sprintf("%s:%d", GetPackageFile(f.File), f.Line)
		},
		PrettyPrint: false,
	}})
	logger.SetOutput(ioutil.Discard)
	if err := logger.ReadLevel(Sys, c); err != nil {
		panic(err)
	}
	if err := logger.ReadHooks(Opt, c, nil); err != nil {
		panic(err)
	}
	return logger
}

type CustomOptFormatter struct {
	logrus.JSONFormatter
}

func (f *CustomOptFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	bytes, ok := entry.Data["Bytes"]
	if ok {
		return bytes.([]byte), nil
	}

	bytes, err := f.JSONFormatter.Format(entry)
	if err != nil {
		return nil, err
	}
	entry.Data["Bytes"] = bytes

	return bytes.([]byte), nil
}
