package hook

import (
	"io"
	"sync"

	"github.com/mattn/go-colorable"
)

var (
	stdout io.Writer
	stderr io.Writer
	mutex  sync.Mutex
)

func getStdout() io.Writer {
	mutex.Lock()
	defer mutex.Unlock()
	if stdout == nil {
		stdout = colorable.NewColorableStdout()
	}
	return stdout
}

func getStderr() io.Writer {
	mutex.Lock()
	defer mutex.Unlock()
	if stderr == nil {
		stderr = colorable.NewColorableStderr()
	}
	return stderr
}
