// +build !windows,!nacl,!plan9

package hook

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type hookCreaterFunc func(string, *Processor, []byte) (logrus.Hook, error)

var hooks = map[string]hookCreaterFunc{
	"lumberjack": NewLumberjackHook,
	"stdout":     NewStdoutHook,
	"stderr":     NewStderrHook,
	"syslog":     NewSyslogHook,
}

// New return a hook init by yaml config
func New(name string, typ string, processor *Processor,
	config []byte) (logrus.Hook, error) {
	creater, ok := hooks[typ]
	if !ok {
		return nil, fmt.Errorf("no hook %s is found", typ)
	}
	return creater(name, processor, config)
}
