package hook

import (
	"encoding/json"
	"io"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
)

// StdoutConfig stores the configuration of StdoutHook
type StdoutConfig struct {
}

// StdoutHook is for stdout
type StdoutHook struct {
	writer.Hook
	processor *Processor
}

type stdoutWriter struct {
	processor *Processor
	writer    io.Writer
}

func (s *stdoutWriter) Write(p []byte) (n int, err error) {
	if s.processor != nil && s.processor.Handler != nil {
		p = s.processor.Process(p)
	}
	return s.writer.Write(p)
}

// NewStdoutHook creates a new stdout hook
func NewStdoutHook(name string, processor *Processor, config []byte,
) (logrus.Hook, error) {
	var c = &StdoutConfig{}
	if err := json.Unmarshal(config, c); err != nil {
		return nil, err
	}

	w := writer.Hook{
		Writer: &stdoutWriter{
			processor: processor,
			writer:    getStdout(),
		},
		LogLevels: []logrus.Level{
			logrus.InfoLevel,
			logrus.DebugLevel,
		},
	}
	return &StdoutHook{w, processor}, nil
}
