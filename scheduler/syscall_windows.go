//go:build windows
// +build windows

package scheduler

import "syscall"

func init() {
	winmmDLL := syscall.NewLazyDLL("winmm.dll")
	procTimeBeginPeriod := winmmDLL.NewProc("timeBeginPeriod")
	_, _, err := procTimeBeginPeriod.Call(uintptr(1))
	if err != nil && err.Error() != "The operation completed successfully." {
		panic(err)
	}
}
