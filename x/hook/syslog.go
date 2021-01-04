// +build !windows,!nacl,!plan9

package hook

import (
	"encoding/json"
	"fmt"
	"log/syslog"
	"os"

	"github.com/sirupsen/logrus"
)

// SyslogConfig stores the configuration of SyslogHook
type SyslogConfig struct {
	Network string
	Addr    string
	Tag     string
	Level   string
	Levels  []string
}

// SyslogHook to send logs via syslog.
type SyslogHook struct {
	writer    *syslog.Writer
	config    *SyslogConfig
	processor *Processor
	LogLevels []logrus.Level
}

// NewSyslogHook Creates a hook to be added to an instance of logger. This is called with
// `hook, err := NewSyslogHook("udp", "localhost:514", syslog.LOG_DEBUG, "")`
// `if err == nil { log.Hooks.Add(hook) }`
func NewSyslogHook(name string, processor *Processor, config []byte,
) (logrus.Hook, error) {
	var c = &SyslogConfig{}
	if err := json.Unmarshal(config, c); err != nil {
		return nil, err
	}

	var logLevels []logrus.Level
	switch {
	case c.Level != "":
		logLevels = aboveLevel(c.Level)
	case len(c.Levels) > 0:
		logLevels = parseLevels(c.Levels)
	default:
		logLevels = logrus.AllLevels
	}

	writer, err := syslog.Dial(c.Network, c.Addr, syslog.LOG_INFO, c.Tag)
	return &SyslogHook{writer, c, processor, logLevels}, err
}

// Fire is called when a log event is fired.
func (hook *SyslogHook) Fire(entry *logrus.Entry) error {
	line, err := entry.Bytes()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read entry, %v", err)
		return err
	}

	if hook.processor != nil && hook.processor.Handler != nil {
		line = hook.processor.Process(line)
	}

	switch entry.Level {
	case logrus.PanicLevel:
		return hook.writer.Crit(string(line))
	case logrus.FatalLevel:
		return hook.writer.Crit(string(line))
	case logrus.ErrorLevel:
		return hook.writer.Err(string(line))
	case logrus.WarnLevel:
		return hook.writer.Warning(string(line))
	case logrus.InfoLevel:
		return hook.writer.Info(string(line))
	case logrus.DebugLevel, logrus.TraceLevel:
		return hook.writer.Debug(string(line))
	default:
		return nil
	}
}

// Levels returns the available logging levels
func (hook *SyslogHook) Levels() []logrus.Level {
	return hook.LogLevels
}
