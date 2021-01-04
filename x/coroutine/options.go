package coroutine

var (
	expire uint = 10
)

// SetExpire sets expire time for all coroutines
// Must be call before all coroutines start
func SetExpire(n uint) {
	expire = n
}
