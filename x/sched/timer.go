package sched

import (
	"math"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lonng/nano/log"
	"github.com/lonng/nano/session"
	"github.com/lonng/nano/x/virtualtime"
)

const (
	infinite = -1
	//TimerPrecision = time.Second
)

type TimerManager struct {
	incrementID int64            // auto increment id
	timers      map[int64]*Timer // all timers

	muClosingTimer sync.RWMutex
	closingTimer   []int64
	muCreatedTimer sync.RWMutex
	createdTimer   []*Timer
}

type (
	// TimerFunc represents a function which will be called periodically in main
	// logic gorontine.
	TimerFunc func()

	// TimerCondition represents a checker that returns true when cron job needs
	// to execute
	TimerCondition interface {
		Check(now time.Time) bool
	}

	// Timer represents a cron job
	Timer struct {
		id        int64          // timer id
		fn        TimerFunc      // function that execute
		createAt  int64          // timer create time
		interval  time.Duration  // execution interval
		condition TimerCondition // condition to cron job execution
		elapse    int64          // total elapse time
		closed    int32          // is timer closed
		counter   int            // counter
	}
)

// ID returns id of current timer
func (t *Timer) ID() int64 {
	return t.id
}

// Stop turns off a timer. After Stop, fn will not be called forever
func (t *Timer) Stop() {
	if atomic.AddInt32(&t.closed, 1) != 1 {
		return
	}

	t.counter = 0
}

// execute job function with protection
func safecall(id int64, fn TimerFunc) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("Handle timer panic: %+v\n%s", err, debug.Stack())
		}
	}()

	fn()
}

func cron(timerManager *TimerManager) {
	if len(timerManager.createdTimer) > 0 {
		timerManager.muCreatedTimer.Lock()
		for _, t := range timerManager.createdTimer {
			timerManager.timers[t.id] = t
		}
		timerManager.createdTimer = timerManager.createdTimer[:0]
		timerManager.muCreatedTimer.Unlock()
	}

	if len(timerManager.timers) < 1 {
		return
	}

	now := virtualtime.Now()
	unn := now.UnixNano()
	for id, t := range timerManager.timers {
		if t.counter == infinite || t.counter > 0 {
			// condition timer
			if t.condition != nil {
				if t.condition.Check(now) {
					safecall(id, t.fn)
				}
				continue
			}

			// execute job
			if t.createAt+t.elapse <= unn {
				safecall(id, t.fn)
				t.elapse += int64(t.interval)

				// update timer counter
				if t.counter != infinite && t.counter > 0 {
					t.counter--
				}
			}
		}

		if t.counter == 0 {
			timerManager.muClosingTimer.Lock()
			timerManager.closingTimer = append(timerManager.closingTimer, t.id)
			timerManager.muClosingTimer.Unlock()
			continue
		}
	}

	if len(timerManager.closingTimer) > 0 {
		timerManager.muClosingTimer.Lock()
		for _, id := range timerManager.closingTimer {
			delete(timerManager.timers, id)
		}
		timerManager.closingTimer = timerManager.closingTimer[:0]
		timerManager.muClosingTimer.Unlock()
	}
}

// NewTimer returns a new Timer containing a function that will be called
// with a period specified by the duration argument. It adjusts the intervals
// for slow receivers.
// The duration d must be greater than zero; if not, NewTimer will panic.
// Stop the timer to release associated resources.
func NewTimer(s *session.Session, interval time.Duration, fn TimerFunc) *Timer {
	return NewCountTimer(s, interval, infinite, fn)
}

// NewCountTimer returns a new Timer containing a function that will be called
// with a period specified by the duration argument. After count times, timer
// will be stopped automatically, It adjusts the intervals for slow receivers.
// The duration d must be greater than zero; if not, NewCountTimer will panic.
// Stop the timer to release associated resources.
func NewCountTimer(s *session.Session, interval time.Duration, count int, fn TimerFunc) *Timer {
	if fn == nil {
		panic("nano/timer: nil timer function")
	}
	if interval <= 0 {
		panic("non-positive interval for NewTimer")
	}

	ss := GetSessionScheduler(s)
	timerManager := ss.timerManager

	t := &Timer{
		id:       atomic.AddInt64(&timerManager.incrementID, 1),
		fn:       fn,
		createAt: virtualtime.Now().UnixNano(),
		interval: interval,
		elapse:   int64(interval), // first execution will be after interval
		counter:  count,
	}

	timerManager.muCreatedTimer.Lock()
	timerManager.createdTimer = append(timerManager.createdTimer, t)
	timerManager.muCreatedTimer.Unlock()
	return t
}

// NewAfterTimer returns a new Timer containing a function that will be called
// after duration that specified by the duration argument.
// The duration d must be greater than zero; if not, NewAfterTimer will panic.
// Stop the timer to release associated resources.
func NewAfterTimer(s *session.Session, duration time.Duration, fn TimerFunc) *Timer {
	return NewCountTimer(s, duration, 1, fn)
}

// NewCondTimer returns a new Timer containing a function that will be called
// when condition satisfied that specified by the condition argument.
// The duration d must be greater than zero; if not, NewCondTimer will panic.
// Stop the timer to release associated resources.
func NewCondTimer(s *session.Session, condition TimerCondition, fn TimerFunc) *Timer {
	if condition == nil {
		panic("nano/timer: nil condition")
	}

	t := NewCountTimer(s, time.Duration(math.MaxInt64), infinite, fn)
	t.condition = condition

	return t
}
