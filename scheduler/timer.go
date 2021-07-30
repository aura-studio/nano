// Copyright (c) nano Authors. All Rights Reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package scheduler

import (
	"math"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aura-studio/nano/log"
)

const (
	infinite = -1
)

type (
	// TimerFunc represents a function which will be called periodically in main
	// logic gorontine.
	TimerFunc func()

	// TimerCondition represents a checker that returns true when cron job needs
	// to execute
	TimerCondition interface {
		Check(now time.Time) bool
	}
)

// Timer represents a cron job
type Timer struct {
	id        int64          // timer id
	fn        TimerFunc      // function that execute
	createAt  int64          // timer create time
	interval  time.Duration  // execution interval
	condition TimerCondition // condition to cron job execution
	elapse    int64          // total elapse time
	closed    int32          // is timer closed
	counter   int            // counter
}

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

type TimerManager struct {
	incrementID int64            // auto increment id
	timers      map[int64]*Timer // all timers

	muClosingTimer sync.RWMutex
	closingTimer   []int64
	muCreatedTimer sync.RWMutex
	createdTimer   []*Timer
}

var (
	// timerManager manager for all timers
	timerManager *TimerManager
)

func init() {
	timerManager = NewTimerManager()
}

func NewTimerManager() *TimerManager {
	return &TimerManager{
		timers: make(map[int64]*Timer),
	}
}

// execute job function with protection
func (tm *TimerManager) safecall(id int64, fn TimerFunc) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("Handle timer panic: %+v\n%s", err, debug.Stack())
		}
	}()

	fn()
}

func (tm *TimerManager) Cron() {
	if len(tm.createdTimer) > 0 {
		tm.muCreatedTimer.Lock()
		for _, t := range tm.createdTimer {
			tm.timers[t.id] = t
		}
		tm.createdTimer = tm.createdTimer[:0]
		tm.muCreatedTimer.Unlock()
	}

	if len(tm.timers) < 1 {
		return
	}

	now := time.Now()
	unn := now.UnixNano()
	for id, t := range tm.timers {
		if t.counter == infinite || t.counter > 0 {
			// condition timer
			if t.condition != nil {
				if t.condition.Check(now) {
					tm.safecall(id, t.fn)
				}
				continue
			}

			// execute job
			if t.createAt+t.elapse <= unn {
				tm.safecall(id, t.fn)
				t.elapse += int64(t.interval)

				// update timer counter
				if t.counter != infinite && t.counter > 0 {
					t.counter--
				}
			}
		}

		if t.counter == 0 {
			tm.muClosingTimer.Lock()
			tm.closingTimer = append(tm.closingTimer, t.id)
			tm.muClosingTimer.Unlock()
			continue
		}
	}

	if len(tm.closingTimer) > 0 {
		tm.muClosingTimer.Lock()
		for _, id := range tm.closingTimer {
			delete(tm.timers, id)
		}
		tm.closingTimer = tm.closingTimer[:0]
		tm.muClosingTimer.Unlock()
	}
}

// NewCountTimer returns a new Timer containing a function that will be called
// with a period specified by the duration argument. After count times, timer
// will be stopped automatically, It adjusts the intervals for slow receivers.
// The duration d must be greater than zero; if not, NewCountTimer will panic.
// Stop the timer to release associated resources.
func (tm *TimerManager) NewCountTimer(interval time.Duration, count int, fn TimerFunc) *Timer {
	if fn == nil {
		panic("nano/timer: nil timer function")
	}
	if interval <= 0 {
		panic("non-positive interval for NewTimer")
	}

	t := &Timer{
		id:       atomic.AddInt64(&tm.incrementID, 1),
		fn:       fn,
		createAt: time.Now().UnixNano(),
		interval: interval,
		elapse:   int64(interval), // first execution will be after interval
		counter:  count,
	}

	tm.muCreatedTimer.Lock()
	tm.createdTimer = append(tm.createdTimer, t)
	tm.muCreatedTimer.Unlock()
	return t
}

// NewAfterTimer returns a new Timer containing a function that will be called
// after duration that specified by the duration argument.
// The duration d must be greater than zero; if not, NewAfterTimer will panic.
// Stop the timer to release associated resources.
func (tm *TimerManager) NewAfterTimer(duration time.Duration, fn TimerFunc) *Timer {
	return tm.NewCountTimer(duration, 1, fn)
}

// NewCondTimer returns a new Timer containing a function that will be called
// when condition satisfied that specified by the condition argument.
// The duration d must be greater than zero; if not, NewCondTimer will panic.
// Stop the timer to release associated resources.
func (tm *TimerManager) NewCondTimer(condition TimerCondition, fn TimerFunc) *Timer {
	if condition == nil {
		panic("nano/timer: nil condition")
	}

	t := tm.NewCountTimer(time.Duration(math.MaxInt64), infinite, fn)
	t.condition = condition

	return t
}

// NewTimer returns a new Timer containing a function that will be called
// with a period specified by the duration argument. It adjusts the intervals
// for slow receivers.
// The duration d must be greater than zero; if not, NewTimer will panic.
// Stop the timer to release associated resources.
func (tm *TimerManager) NewTimer(interval time.Duration, fn TimerFunc) *Timer {
	return tm.NewCountTimer(interval, infinite, fn)
}
