package hook

import (
	"encoding/json"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"

	"fmt"

	"github.com/sirupsen/logrus"
)

// LumberjackConfig stores the configuration of LumberjackHook
type LumberjackConfig struct {
	File      string
	Size      int
	Backup    int
	Day       int
	Compress  bool
	Directory string
}

// LumberjackHook stores the hook of rolling file appender
type LumberjackHook struct {
	logger    *lumberjack.Logger
	config    *LumberjackConfig
	processor *Processor
}

// NewLumberjackHook creates a new LumberjackHook
func NewLumberjackHook(name string, processor *Processor, config []byte,
) (logrus.Hook, error) {
	var c = &LumberjackConfig{}
	if err := json.Unmarshal(config, c); err != nil {
		return nil, err
	}

	logger := &lumberjack.Logger{
		Filename:   filepath.Join(logPath, c.File),
		MaxSize:    c.Size, // megabytes
		MaxBackups: c.Backup,
		MaxAge:     c.Day,      //days
		Compress:   c.Compress, // disabled by default
	}
	if logger == nil {
		return nil, fmt.Errorf("lumberjack logger is nil")
	}

	return &LumberjackHook{logger, c, processor}, nil
}

// Fire is called when a log event is fired.
func (hook *LumberjackHook) Fire(entry *logrus.Entry) error {
	// Convert the line to string
	line, err := entry.Bytes()
	if err != nil {
		return err
	}

	if hook.processor != nil && hook.processor.Handler != nil {
		line = hook.processor.Process(line)
	}
	// Write the the logger
	_, err = hook.logger.Write(line)
	if err != nil {
		return err
	}

	return nil
}

// Levels returns the available logging levels
func (hook *LumberjackHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
