package hook

// func TestStdout(t *testing.T) {
// 	p := &Processor{
// 		StringHandler: func(s string) string {
// 			return "[TestStdout]" + s
// 		},
// 		BytesHandler: func(s []byte) []byte {
// 			var header = []byte("[TestStdout]")
// 			return append(header, s...)
// 		},
// 	}
// 	hook, err := New("run", "stdout", p)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	logger := logrus.New()
// 	logger.SetReportCaller(true)
// 	logger.SetFormatter(&logrus.TextFormatter{
// 		ForceColors:            true,
// 		TimestampFormat:        "2006/02/01 15:04:05.0000000", // the "time" field configuratiom
// 		FullTimestamp:          true,
// 		DisableLevelTruncation: true, // log level field configuration
// 		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
// 			return "", fmt.Sprintf(" %s:%d:", getPackageFile(f.File), f.Line)
// 		},
// 	})
// 	logger.SetOutput(ioutil.Discard) // Send all logs to nowhere by default
// 	logger.Hooks.Add(hook)
// 	logger.Println("Test stdout one line")
// }

// func TestStderr(t *testing.T) {
// 	p := &Processor{
// 		StringHandler: func(s string) string {
// 			return "[TestStderr]" + s
// 		},
// 		BytesHandler: func(s []byte) []byte {
// 			var header = []byte("[TestStderr]")
// 			return append(header, s...)
// 		},
// 	}
// 	hook, err := New("run", "stderr", p)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	logger := logrus.New()
// 	logger.SetReportCaller(true)
// 	logger.SetFormatter(&logrus.TextFormatter{
// 		ForceColors:            true,
// 		TimestampFormat:        "2006/02/01 15:04:05.0000000", // the "time" field configuratiom
// 		FullTimestamp:          true,
// 		DisableLevelTruncation: true, // log level field configuration
// 		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
// 			return "", fmt.Sprintf(" %s:%d:", getPackageFile(f.File), f.Line)
// 		},
// 	})
// 	logger.SetOutput(ioutil.Discard) // Send all logs to nowhere by default
// 	logger.Hooks.Add(hook)
// 	logger.Warnln("Test stderr one line")
// }

// func TestLumberJack(t *testing.T) {
// 	p := &Processor{
// 		StringHandler: func(s string) string {
// 			return "[TestLumberJack]" + s
// 		},
// 		BytesHandler: func(s []byte) []byte {
// 			var header = []byte("[TestLumberJack]")
// 			return append(header, s...)
// 		},
// 	}
// 	hook, err := New("run", "lumberjack", p)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	logger := logrus.New()
// 	logger.SetReportCaller(true)
// 	logger.SetFormatter(&logrus.TextFormatter{
// 		ForceColors:            true,
// 		TimestampFormat:        "2006/02/01 15:04:05.0000000", // the "time" field configuratiom
// 		FullTimestamp:          true,
// 		DisableLevelTruncation: true, // log level field configuration
// 		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
// 			return "", fmt.Sprintf(" %s:%d:", getPackageFile(f.File), f.Line)
// 		},
// 	})
// 	logger.SetOutput(ioutil.Discard) // Send all logs to nowhere by default
// 	logger.Hooks.Add(hook)
// 	logger.Println("Test lumberjack one line")
// }
