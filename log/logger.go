// Copyright (c) nano Authors. All Rights Reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package log

import (
	"log"
	"os"
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

func init() {
	Use(log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile))
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

// Use rewrites the default logger
func Use(logger Logger) {
	if logger == nil {
		return
	}

	_, ok := logger.(LevelLogger)
	if !ok {
		setLogger(logger)
	} else {
		setLevelLogger(logger.(LevelLogger))
	}
}

func setLogger(logger Logger) {
	Tracef = logger.Printf
	Trace = logger.Print
	Traceln = logger.Println

	Debugf = logger.Printf
	Debug = logger.Print
	Debugln = logger.Println

	Infof = logger.Printf
	Info = logger.Print
	Infoln = logger.Println

	Printf = logger.Printf
	Print = logger.Print
	Println = logger.Println

	Warnf = logger.Printf
	Warn = logger.Print
	Warnln = logger.Println

	Warningf = logger.Printf
	Warning = logger.Print
	Warningln = logger.Println

	Errorf = logger.Printf
	Error = logger.Print
	Errorln = logger.Println

	Fatalf = logger.Fatalf
	Fatal = logger.Fatal
	Fatalln = logger.Fatalln

	Panicf = logger.Panicf
	Panic = logger.Panic
	Panicln = logger.Panicln
}

func setLevelLogger(logger LevelLogger) {
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
