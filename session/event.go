package session

import (
	"runtime/debug"

	"github.com/aura-studio/nano/log"
)

var (
	onInited     []func(s *Session) // call func in slice when session is inited
	afterInited  []func(s *Session) // call func in slice after session is inited
	beforeClosed []func(s *Session) // call func in slice before session is closed
	onClosed     []func(s *Session) // call func in slice when session is closed
)

// OnInited set a func that will be called on session inited
func OnInited(f func(*Session)) {
	onInited = append(onInited, f)
}

// AfterInited set a func that will be called after session inited
func AfterInited(f func(*Session)) {
	afterInited = append(afterInited, f)
}

// BeforeClosed set a func that will be called before session closed
func BeforeClosed(f func(*Session)) {
	beforeClosed = append(beforeClosed, f)
}

// OnClosed set a func that will be called on session closed
func OnClosed(f func(*Session)) {
	onClosed = append(onClosed, f)
}

// Inited call all funcs that was registered by OnInited
func Inited(s *Session) {
	for _, f := range onInited {
		SafeCall(func() {
			f(s)
		})
	}
	for _, f := range afterInited {
		SafeCall(func() {
			f(s)
		})
	}
}

// Closed call all funcs that was registered by OnClosed
func Closed(s *Session) {
	for _, f := range beforeClosed {
		SafeCall(func() {
			f(s)
		})
	}
	for _, f := range onClosed {
		SafeCall(func() {
			f(s)
		})
	}
}

func SafeCall(f func()) {
	defer func() {
		if v := recover(); v != nil {
			if err, ok := v.(error); ok {
				log.Errorf("panic: %v\n%s", err, string(debug.Stack()))
			} else {
				log.Errorf("panic: %v\n%s", v, string(debug.Stack()))
			}
		}
	}()

	f()
}
