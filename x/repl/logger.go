package repl

import (
	"gopkg.in/abiosoft/ishell.v2"
	l "log"
	"os"
)

// Logger comment
type Logger interface {
	Printf(format string, v ...interface{})
	Print(v ...interface{})
	Println(v ...interface{})
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
