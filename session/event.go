package session

import (
	"runtime/debug"
	"sort"

	"github.com/aura-studio/nano/log"
)

var (
	onCreated           = make(map[int][]func(s *Session)) // call func in slice when session is created
	onCreatedPriorities []int

	afterCreated           = make(map[int][]func(s *Session)) // call func in slice after session is created
	afterCreatedPriorities []int

	onInited           = make(map[int][]func(s *Session)) // call func in slice when session is inited
	onInitedPriorities []int

	afterInited           = make(map[int][]func(s *Session)) // call func in slice after session is inited
	afterInitedPriorities []int

	beforeClosed           = make(map[int][]func(s *Session)) // call func in slice before session is closed
	beforeClosedPriorities []int

	onClosed           = make(map[int][]func(s *Session)) // call func in slice when session is closed
	onClosedPriorities []int
)

// OnCreated set a func that will be called on session created
func OnCreated(f func(*Session), p ...int) {
	var priority int
	if len(p) > 0 {
		priority = p[0]
	}
	onCreated[priority] = append(onCreated[priority], f)

	onCreatedPriorities = make([]int, 0, len(onCreated))
	for k := range onCreated {
		onCreatedPriorities = append(onCreatedPriorities, k)
	}
	sort.Ints(onCreatedPriorities)
}

// AfterCreated set a func that will be called after session created
func AfterCreated(f func(*Session), p ...int) {
	var priority int
	if len(p) > 0 {
		priority = p[0]
	}
	afterCreated[priority] = append(afterCreated[priority], f)

	afterCreatedPriorities = make([]int, 0, len(afterCreated))
	for k := range afterCreated {
		afterCreatedPriorities = append(afterCreatedPriorities, k)
	}
	sort.Ints(afterCreatedPriorities)
}

// OnInited set a func that will be called on session inited
func OnInited(f func(*Session), p ...int) {
	var priority int
	if len(p) > 0 {
		priority = p[0]
	}
	onInited[priority] = append(onInited[priority], f)

	onInitedPriorities = make([]int, 0, len(onInited))
	for k := range onInited {
		onInitedPriorities = append(onInitedPriorities, k)
	}
	sort.Ints(onInitedPriorities)
}

// AfterInited set a func that will be called after session inited
func AfterInited(f func(*Session), p ...int) {
	var priority int
	if len(p) > 0 {
		priority = p[0]
	}
	afterInited[priority] = append(afterInited[priority], f)

	afterInitedPriorities = make([]int, 0, len(afterInited))
	for k := range afterInited {
		afterInitedPriorities = append(afterInitedPriorities, k)
	}
	sort.Ints(afterInitedPriorities)
}

// BeforeClosed set a func that will be called before session closed
func BeforeClosed(f func(*Session), p ...int) {
	var priority int
	if len(p) > 0 {
		priority = p[0]
	}
	beforeClosed[priority] = append(beforeClosed[priority], f)

	beforeClosedPriorities = make([]int, 0, len(beforeClosed))
	for k := range beforeClosed {
		beforeClosedPriorities = append(beforeClosedPriorities, k)
	}
	sort.Ints(beforeClosedPriorities)
}

// OnClosed set a func that will be called on session closed
func OnClosed(f func(*Session), p ...int) {
	var priority int
	if len(p) > 0 {
		priority = p[0]
	}
	onClosed[priority] = append(onClosed[priority], f)

	onClosedPriorities = make([]int, 0, len(onClosed))
	for k := range onClosed {
		onClosedPriorities = append(onClosedPriorities, k)
	}
	sort.Ints(onClosedPriorities)
}

func Created(s *Session) {
	for _, priority := range onCreatedPriorities {
		for _, f := range onCreated[priority] {
			SafeCall(func() {
				f(s)
			})
		}
	}
	for _, priority := range afterCreatedPriorities {
		for _, f := range afterCreated[priority] {
			SafeCall(func() {
				f(s)
			})
		}
	}
}

// Inited call all funcs that was registered by OnInited
func Inited(s *Session) {
	for _, priority := range onInitedPriorities {
		for _, f := range onInited[priority] {
			SafeCall(func() {
				f(s)
			})
		}
	}
	for _, priority := range afterInitedPriorities {
		for _, f := range afterInited[priority] {
			SafeCall(func() {
				f(s)
			})
		}
	}
}

// Closed call all funcs that was registered by OnClosed
func Closed(s *Session) {
	for _, priority := range beforeClosedPriorities {
		for _, f := range beforeClosed[priority] {
			SafeCall(func() {
				f(s)
			})
		}
	}

	for _, priority := range onClosedPriorities {
		for _, f := range onClosed[priority] {
			SafeCall(func() {
				f(s)
			})
		}
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
