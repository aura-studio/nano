package log

import (
	"log"
)

// Logger represents the log interface
type Logger interface {
	Printf(format string, v ...interface{})
	Print(v ...interface{})
	Println(v ...interface{})

	Fatalf(format string, v ...interface{})
	Fatal(v ...interface{})
	Fatalln(v ...interface{})

	Panicf(format string, v ...interface{})
	Panic(v ...interface{})
	Panicln(v ...interface{})
}

// LevelLogger represents the level log interface
type LevelLogger interface {
	Tracef(format string, v ...interface{})
	Trace(v ...interface{})
	Traceln(v ...interface{})

	Debugf(format string, v ...interface{})
	Debug(v ...interface{})
	Debugln(v ...interface{})

	Infof(format string, v ...interface{})
	Info(v ...interface{})
	Infoln(v ...interface{})

	Printf(format string, v ...interface{})
	Print(v ...interface{})
	Println(v ...interface{})

	Warnf(format string, v ...interface{})
	Warn(v ...interface{})
	Warnln(v ...interface{})

	Warningf(format string, v ...interface{})
	Warning(v ...interface{})
	Warningln(v ...interface{})

	Errorf(format string, v ...interface{})
	Error(v ...interface{})
	Errorln(v ...interface{})

	Fatalf(format string, v ...interface{})
	Fatal(v ...interface{})
	Fatalln(v ...interface{})

	Panicf(format string, v ...interface{})
	Panic(v ...interface{})
	Panicln(v ...interface{})
}

// log
var (
	Printf  func(format string, v ...interface{})
	Print   func(v ...interface{})
	Println func(v ...interface{})

	Fatalf  func(format string, v ...interface{})
	Fatal   func(v ...interface{})
	Fatalln func(v ...interface{})

	Panicf  func(format string, v ...interface{})
	Panic   func(v ...interface{})
	Panicln func(v ...interface{})

	Tracef  func(format string, v ...interface{})
	Trace   func(v ...interface{})
	Traceln func(v ...interface{})

	Debugf  func(format string, v ...interface{})
	Debug   func(v ...interface{})
	Debugln func(v ...interface{})

	Infof  func(format string, v ...interface{})
	Info   func(v ...interface{})
	Infoln func(v ...interface{})

	Warnf  func(format string, v ...interface{})
	Warn   func(v ...interface{})
	Warnln func(v ...interface{})

	Warningf  func(format string, v ...interface{})
	Warning   func(v ...interface{})
	Warningln func(v ...interface{})

	Errorf  func(format string, v ...interface{})
	Error   func(v ...interface{})
	Errorln func(v ...interface{})
)

func init() {
	Tracef = log.Printf
	Trace = log.Print
	Traceln = log.Println

	Debugf = log.Printf
	Debug = log.Print
	Debugln = log.Println

	Infof = log.Printf
	Info = log.Print
	Infoln = log.Println

	Printf = log.Printf
	Print = log.Print
	Println = log.Println

	Warnf = log.Printf
	Warn = log.Print
	Warnln = log.Println

	Warningf = log.Printf
	Warning = log.Print
	Warningln = log.Println

	Errorf = log.Printf
	Error = log.Print
	Errorln = log.Println

	Fatalf = log.Fatalf
	Fatal = log.Fatal
	Fatalln = log.Fatalln

	Panicf = log.Panicf
	Panic = log.Panic
	Panicln = log.Panicln
}

func SetLogger(logger LevelLogger) {
	Tracef = logger.Tracef
	Trace = logger.Trace
	Traceln = logger.Traceln

	Debugf = logger.Debugf
	Debug = logger.Debug
	Debugln = logger.Debugln

	Infof = logger.Infof
	Info = logger.Info
	Infoln = logger.Infoln

	Printf = logger.Printf
	Print = logger.Print
	Println = logger.Println

	Warnf = logger.Warnf
	Warn = logger.Warn
	Warnln = logger.Warnln

	Warningf = logger.Warningf
	Warning = logger.Warning
	Warningln = logger.Warningln

	Errorf = logger.Errorf
	Error = logger.Error
	Errorln = logger.Errorln

	Fatalf = logger.Fatalf
	Fatal = logger.Fatal
	Fatalln = logger.Fatalln

	Panicf = logger.Panicf
	Panic = logger.Panic
	Panicln = logger.Panicln
}
