package session

var (
	onOpened []func(s *Session) // call func in slice when session is opened
	onClosed []func(s *Session) // call func in slice when session is closed
)

// OnOpened set a func that will be called on session opened
func OnOpened(f func(*Session)) {
	onOpened = append(onOpened, f)
}

// Opened call all funcs that was registerd by OnOpened
func Opened(s *Session) {
	for _, f := range onOpened {
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
