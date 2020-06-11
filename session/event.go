package session

var (
	onInited []func(s *Session) // call func in slice when session is inited
	onClosed []func(s *Session) // call func in slice when session is closed
)

// OnInited set a func that will be called on session inited
func OnInited(f func(*Session)) {
	onInited = append(onInited, f)
}

// Inited call all funcs that was registerd by OnInited
func Inited(s *Session) {
	for _, f := range onInited {
		f(s)
	}
}

// OnClosed set a func that will be called on session closed
func OnClosed(f func(*Session)) {
	onClosed = append(onClosed, f)
}

// Closed call all funcs that was registerd by OnClosed
func Closed(s *Session) {
	for _, f := range onClosed {
		f(s)
	}
}
