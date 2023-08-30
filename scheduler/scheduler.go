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

	"github.com/aura-studio/nano/env"
	"github.com/aura-studio/nano/log"
	"github.com/aura-studio/nano/session"
)

type (
	// Task is the unit to be scheduled
	Task func()

	scheduler struct {
		TimerManager
		chDie     chan struct{}
		chExit    chan struct{}
		chTasks   chan Task
		closeOnce sync.Once
		running   atomic.Bool
	}

	Context struct {
		ServiceName string
		HandlerName string
		Data        interface{}
	}

	ScheduleFunc func(*session.Session, *Context, Task)

	Scheduler interface {
		TimerManager
		Schedule(*session.Session, *Context, Task)
		PushTask(task Task)
		Digest()
		Close()
	}
)

var (
	global Scheduler
)

func init() {
	global = NewScheduler()
}

func Global() Scheduler {
	return global
}

// NewScheduler creates a new TimerScheduler
func NewScheduler() *scheduler {
	s := &scheduler{
		TimerManager: NewTimerManager(),
		chDie:        make(chan struct{}),
		chExit:       make(chan struct{}),
		chTasks:      make(chan Task, 1<<8),
	}

	go func() {
		defer func() {
			if v := recover(); v != nil {
				if err, ok := v.(error); ok {
					log.Errorf("panic: %v\n%s", err, string(debug.Stack()))
				} else {
					log.Errorf("panic: %v\n%s", v, string(debug.Stack()))
				}
			}
		}()

		s.Digest()
	}()

	s.running.Store(true)

	return s
}

func Digest() {
	global.Digest()
}

func (s *scheduler) Digest() {
	defer func() {
		s.TimerManager.CloseTimer()
		close(s.chTasks)
		close(s.chExit)
	}()

	for {
		select {
		case f := <-s.TimerManager.TaskChan():
			func() {
				defer func() {
					if err := recover(); err != nil {
						log.Errorf("panic: %v\n%s", err.(error), string(debug.Stack()))
					}
				}()

				f()
			}()

		case f := <-s.chTasks:
			func() {
				defer func() {
					if err := recover(); err != nil {
						log.Errorf("panic: %v\n%s", err.(error), string(debug.Stack()))
					}
				}()

				f()
			}()

		case <-s.chDie:
			s.running.Store(false)
			return
		}
	}
}

func Close() {
	global.Close()
}

// Close closes scheduler
func (s *scheduler) Close() {
	s.closeOnce.Do(func() {
		close(s.chDie)
		<-s.chExit
	})
}

func Schedule(_ *session.Session, _ interface{}, task Task) {
	global.PushTask(task)
}

func (s *scheduler) Schedule(_ *session.Session, _ *Context, task Task) {
	s.PushTask(task)
}

func PushTask(task Task) {
	global.PushTask(task)
}

// Schedule implements scheduler.Schedule
func (s *scheduler) PushTask(task Task) {
	if s.running.Load() {
		s.chTasks <- task
	}
}

const (
	Infinite = -1
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

type TimerManager interface {
	NewCountTimer(interval time.Duration, count int, fn TimerFunc) *Timer
	NewAfterTimer(duration time.Duration, fn TimerFunc) *Timer
	NewCondTimer(condition TimerCondition, fn TimerFunc) *Timer
	NewTimer(interval time.Duration, fn TimerFunc) *Timer
	TaskChan() <-chan Task
	CloseTimer()
}

type timerManager struct {
	chDie     chan struct{}
	chExit    chan struct{}
	chTask    chan Task
	closeOnce sync.Once
	initOnce  sync.Once
	running   atomic.Bool

	incrementID int64            // auto increment id
	timers      map[int64]*Timer // all timers

	muClosingTimer sync.RWMutex
	closingTimer   []int64
	muCreatedTimer sync.RWMutex
	createdTimer   []*Timer
}

func NewTimerManager() TimerManager {
	return &timerManager{
		chDie:  make(chan struct{}),
		chExit: make(chan struct{}),
		chTask: make(chan Task, 1<<8),

		timers: make(map[int64]*Timer),
	}
}

func (tm *timerManager) CloseTimer() {
	tm.lazy()

	tm.closeOnce.Do(func() {
		close(tm.chDie)
		<-tm.chExit
	})
}

func (tm *timerManager) lazy() {
	tm.initOnce.Do(func() {
		tm.init()
	})
}

func (tm *timerManager) init() {
	go func() {
		defer func() {
			if v := recover(); v != nil {
				if err, ok := v.(error); ok {
					log.Errorf("panic: %v\n%s", err, string(debug.Stack()))
				} else {
					log.Errorf("panic: %v\n%s", v, string(debug.Stack()))
				}
			}
		}()

		tm.digest()
	}()

	tm.running.Store(true)
}

func (tm *timerManager) digest() {
	ticker := time.NewTicker(env.TimerPrecision)
	defer func() {
		ticker.Stop()
		close(tm.chTask)
		close(tm.chExit)
	}()

	for {
		select {
		case <-ticker.C:
			if tm.running.Load() {
				tm.chTask <- tm.cron
			}
		case <-tm.chDie:
			tm.running.Store(false)
			return
		}
	}
}

func (tm *timerManager) TaskChan() <-chan Task {
	return tm.chTask
}

// execute job function with protection
func (tm *timerManager) safecall(id int64, fn TimerFunc) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("Handle timer %d panic: %+v\n%s", id, err, debug.Stack())
		}
	}()

	fn()
}

func (tm *timerManager) cron() {
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
		if t.counter == Infinite || t.counter > 0 {
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
				if t.counter != Infinite && t.counter > 0 {
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

func NewCountTimer(interval time.Duration, count int, fn TimerFunc) *Timer {
	return global.NewCountTimer(interval, count, fn)
}

// NewCountTimer returns a new Timer containing a function that will be called
// with a period specified by the duration argument. After count times, timer
// will be stopped automatically, It adjusts the intervals for slow receivers.
// The duration d must be greater than zero; if not, NewCountTimer will panic.
// Stop the timer to release associated resources.
func (tm *timerManager) NewCountTimer(interval time.Duration, count int, fn TimerFunc) *Timer {
	tm.lazy()
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

func NewAfterTimer(duration time.Duration, fn TimerFunc) *Timer {
	return global.NewAfterTimer(duration, fn)
}

// NewAfterTimer returns a new Timer containing a function that will be called
// after duration that specified by the duration argument.
// The duration d must be greater than zero; if not, NewAfterTimer will panic.
// Stop the timer to release associated resources.
func (tm *timerManager) NewAfterTimer(duration time.Duration, fn TimerFunc) *Timer {
	return tm.NewCountTimer(duration, 1, fn)
}

func NewCondTimer(condition TimerCondition, fn TimerFunc) *Timer {
	return global.NewCondTimer(condition, fn)
}

// NewCondTimer returns a new Timer containing a function that will be called
// when condition satisfied that specified by the condition argument.
// The duration d must be greater than zero; if not, NewCondTimer will panic.
// Stop the timer to release associated resources.
func (tm *timerManager) NewCondTimer(condition TimerCondition, fn TimerFunc) *Timer {
	if condition == nil {
		panic("nano/timer: nil condition")
	}

	t := tm.NewCountTimer(time.Duration(math.MaxInt64), Infinite, fn)
	t.condition = condition

	return t
}

func NewTimer(interval time.Duration, fn TimerFunc) *Timer {
	return global.NewTimer(interval, fn)
}

// NewTimer returns a new Timer containing a function that will be called
// with a period specified by the duration argument. It adjusts the intervals
// for slow receivers.
// The duration d must be greater than zero; if not, NewTimer will panic.
// Stop the timer to release associated resources.
func (tm *timerManager) NewTimer(interval time.Duration, fn TimerFunc) *Timer {
	return tm.NewCountTimer(interval, Infinite, fn)
}
