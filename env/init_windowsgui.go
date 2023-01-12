//go:build windowsgui
// +build windowsgui

package env

import (
	"fmt"
	"os"
	"syscall"
)

func init() {
	windowsGUIRunInConsole()
}

// WindowsGUIRunInConsole print console logs in windows gui mode
func windowsGUIRunInConsole() {
	modkernel32 := syscall.NewLazyDLL("kernel32.dll")
	procAllocConsole := modkernel32.NewProc("AllocConsole")
	r, _, sysErr := syscall.Syscall(procAllocConsole.Addr(), 0, 0, 0, 0)
	if r == 0 {
		fmt.Printf("Could not allocate console: %s. Check build flags..", sysErr)
		os.Exit(2)
	}
	hOut, err := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	if err != nil {
		fmt.Printf("Could not get stdout handle: %v", err)
		os.Exit(3)
	}
	herr, err := syscall.GetStdHandle(syscall.STD_ERROR_HANDLE)
	if err != nil {
		fmt.Printf("Could not get stderr handle: %v", err)
		os.Exit(3)
	}
	os.Stdout = os.NewFile(uintptr(hOut), "/dev/stdout")
	os.Stderr = os.NewFile(uintptr(herr), "/dev/stderr")
}
