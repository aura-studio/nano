package repl

import (
	l "log"
	"os"

	"github.com/aura-studio/nano/log"
	"gopkg.in/abiosoft/ishell.v2"
)

// Logger comment
type Logger interface {
	Printf(format string, v ...interface{})
	Print(v ...interface{})
	Println(v ...interface{})
}

var logger *CliLogger

func initLogger() {
	logger = NewCliLogger(nil)
	log.Use(logger)
}

// CliLogger log by *ishell.Shell
type CliLogger struct {
	shell Logger
}

// NewCliLogger creates a clilog object pointer
func NewCliLogger(shell *ishell.Shell) *CliLogger {
	if shell == nil {
		return &CliLogger{shell: l.New(os.Stdout, "", 0)}
	}
	return &CliLogger{shell: shell}
}

// Warningf comment
func (l CliLogger) Warningf(format string, v ...interface{}) {
	l.shell.Printf(format, v...)
}

// Warning comment
func (l CliLogger) Warning(v ...interface{}) {
	l.shell.Print(v...)
}

// Warningln comment
func (l CliLogger) Warningln(v ...interface{}) {
	err, ok := v[0].(error)
	if ok {
		l.shell.Println(err)
		s, ok := l.shell.(*ishell.Shell)
		if ok {
			l.shell.Print(">>> ")
			s.ShowPrompt(true)
		}
	} else {
		l.shell.Println(v...)
	}
}

// Warnf comment
func (l CliLogger) Warnf(format string, v ...interface{}) {
	l.shell.Printf(format, v...)
}

// Warn comment
func (l CliLogger) Warn(v ...interface{}) {
	l.shell.Print(v...)
}

// Warnln comment
func (l CliLogger) Warnln(v ...interface{}) {
	err, ok := v[0].(error)
	if ok {
		l.shell.Println(err)
		s, ok := l.shell.(*ishell.Shell)
		if ok {
			l.shell.Print(">>> ")
			s.ShowPrompt(true)
		}
	} else {
		l.shell.Println(v...)
	}
}

// Tracef comment
func (l CliLogger) Tracef(format string, v ...interface{}) {
	l.shell.Printf(format, v...)
}

// Trace comment
func (l CliLogger) Trace(v ...interface{}) {
	l.shell.Print(v...)
}

// Traceln comment
func (l CliLogger) Traceln(v ...interface{}) {
	err, ok := v[0].(error)
	if ok {
		l.shell.Println(err)
		s, ok := l.shell.(*ishell.Shell)
		if ok {
			l.shell.Print(">>> ")
			s.ShowPrompt(true)
		}
	} else {
		l.shell.Println(v...)
	}
}

// Debugf comment
func (l CliLogger) Debugf(format string, v ...interface{}) {
	l.shell.Printf(format, v...)
}

// Debug comment
func (l CliLogger) Debug(v ...interface{}) {
	l.shell.Print(v...)
}

// Debugln comment
func (l CliLogger) Debugln(v ...interface{}) {
	err, ok := v[0].(error)
	if ok {
		l.shell.Println(err)
		s, ok := l.shell.(*ishell.Shell)
		if ok {
			l.shell.Print(">>> ")
			s.ShowPrompt(true)
		}
	} else {
		l.shell.Println(v...)
	}
}

// Errorf comment
func (l CliLogger) Errorf(format string, v ...interface{}) {
	l.shell.Printf(format, v...)
}

// Error comment
func (l CliLogger) Error(v ...interface{}) {
	l.shell.Print(v...)
}

// Errorln comment
func (l CliLogger) Errorln(v ...interface{}) {
	err, ok := v[0].(error)
	if ok {
		l.shell.Println(err)
		s, ok := l.shell.(*ishell.Shell)
		if ok {
			l.shell.Print(">>> ")
			s.ShowPrompt(true)
		}
	} else {
		l.shell.Println(v...)
	}
}

// Infof comment
func (l CliLogger) Infof(format string, v ...interface{}) {
	l.shell.Printf(format, v...)
}

// Info comment
func (l CliLogger) Info(v ...interface{}) {
	l.shell.Print(v...)
}

// Infoln comment
func (l CliLogger) Infoln(v ...interface{}) {
	err, ok := v[0].(error)
	if ok {
		l.shell.Println(err)
		s, ok := l.shell.(*ishell.Shell)
		if ok {
			l.shell.Print(">>> ")
			s.ShowPrompt(true)
		}
	} else {
		l.shell.Println(v...)
	}
}

// Printf comment
func (l CliLogger) Printf(format string, v ...interface{}) {
	l.shell.Printf(format, v...)
}

// Print comment
func (l CliLogger) Print(v ...interface{}) {
	l.shell.Print(v...)
}

// Println comment
func (l CliLogger) Println(v ...interface{}) {
	err, ok := v[0].(error)
	if ok {
		l.shell.Println(err)
		s, ok := l.shell.(*ishell.Shell)
		if ok {
			l.shell.Print(">>> ")
			s.ShowPrompt(true)
		}
	} else {
		l.shell.Println(v...)
	}
}

// Fatalf comment
func (l CliLogger) Fatalf(format string, v ...interface{}) {
	l.shell.Printf(format, v...)
}

// Fatal comment
func (l CliLogger) Fatal(v ...interface{}) {
	l.shell.Print(v...)
}

// Fatalln comment
func (l CliLogger) Fatalln(v ...interface{}) {
	l.shell.Println(v...)
}

// Panicf comment
func (l CliLogger) Panicf(format string, v ...interface{}) {
	l.shell.Printf(format, v...)
}

// Panic comment
func (l CliLogger) Panic(v ...interface{}) {
	l.shell.Print(v...)
}

// Panicln comment
func (l CliLogger) Panicln(v ...interface{}) {
	l.shell.Println(v...)
}
