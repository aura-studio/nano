package logger

import (
	"fmt"
	"io/ioutil"
	"runtime"

	"github.com/lonng/nano/x/hook"
	"github.com/lonng/nano/x/virtualtime"

	"github.com/sirupsen/logrus"
)

// NewRunLogger creates a logger hooked by lumberjack.Logger,
// which produce logs that are stored in aliyun's sls production.
// It is for common use, such as "log.Println", and shows in stderr.
func NewRunLogger(c map[string]interface{}) *Logger {
	logger := NewLogger()
	logger.SetReportCaller(true)
	logger.SetFormatter(&CustomRunFormatter{logrus.TextFormatter{
		ForceColors:            true,
		TimestampFormat:        "2006/01/02 15:04:05.0000000", // the "time" field configuratiom
		FullTimestamp:          true,
		DisableLevelTruncation: true, // log upgrade field configuration
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return "", fmt.Sprintf(" %s:%d:", GetPackageFile(f.File), f.Line)
		},
	}})
	logger.SetOutput(ioutil.Discard)
	if err := logger.Hook(Run, c, map[string]*hook.Processor{
		"lumberjack": {
			Handler: func(s []byte) []byte {
				return s[13:]
			},
		},
	}); err != nil {
		panic(err)
	}
	return logger
}

type CustomRunFormatter struct {
	logrus.TextFormatter
}

func (f *CustomRunFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	bytes, ok := entry.Data["Bytes"]
	if ok {
		return bytes.([]byte), nil
	}

	entry.Time = virtualtime.Now()
	bytes, err := f.TextFormatter.Format(entry)
	if err != nil {
		return nil, err
	}

	entry.Data["Bytes"] = bytes
	return bytes.([]byte), nil
}
