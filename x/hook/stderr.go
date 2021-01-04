package hook

import (
	"encoding/json"
	"io"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
)

// StderrConfig stores the configuration of StderrHook
type StderrConfig struct {
}

// StderrHook is for stdout
type StderrHook struct {
	writer.Hook
	processor *Processor
}

type stderrWriter struct {
	processor *Processor
	writer    io.Writer
}

func (s *stderrWriter) Write(p []byte) (n int, err error) {
	if s.processor != nil && s.processor.Handler != nil {
		p = s.processor.Process(p)
	}
	return s.writer.Write(p)
}

// NewStderrHook creates a new stdout hook
func NewStderrHook(name string, processor *Processor, config []byte,
) (logrus.Hook, error) {
	var c = &StderrConfig{}
	if err := json.Unmarshal(config, c); err != nil {
		return nil, err
	}

	w := writer.Hook{
		Writer: &stderrWriter{
			processor: processor,
			writer:    getStderr(),
		},
		LogLevels: []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
		},
	}
	return &StderrHook{w, processor}, nil
}
