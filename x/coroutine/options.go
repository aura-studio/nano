package coroutine

var (
	expire uint = 30
)

// SetExpire sets expire time for all coroutines
// Must be call before all coroutines start
func SetExpire(n uint) {
	expire = n
}
