package hook

var logPath string

// SetLogPath sets log path for jumberjack
// Must be called before all lumberjack logs start
func SetLogPath(s string) {
	logPath = s
}
