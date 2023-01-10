//go:build windows
// +build windows

package scheduler

import "syscall"

func init() {
	winmmDLL := syscall.NewLazyDLL("winmm.dll")
	procTimeBeginPeriod := winmmDLL.NewProc("timeBeginPeriod")
	procTimeBeginPeriod.Call(uintptr(1))
}
