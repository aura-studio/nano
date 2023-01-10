//go:build windows
// +build windows

package scheduler

import "syscall"

func init() {
	winmmDLL := syscall.NewLazyDLL("winmm.dll")
	procTimeBeginPeriod := winmmDLL.NewProc("timeBeginPeriod")
	if _, _, err := procTimeBeginPeriod.Call(uintptr(1)); err != nil {
		panic(err)
	}
}
