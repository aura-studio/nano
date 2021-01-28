package sched

import (
	"runtime/debug"
	"sync/atomic"

	"github.com/lonng/nano/x/log"
	"github.com/lonng/nano/scheduler"
	"github.com/lonng/nano/session"
)

type (
	// HashScheduler implements scheduler.LocalScheduler
	HashScheduler struct {
		chDie    chan struct{}
		chExit   chan struct{}
		chTasks  chan scheduler.Task
		started  int32
		closed   int32
		schedTag string
		index    uint
	}
)

// NewHashScheduler creates a new HashScheduler
func NewHashScheduler(tag string, index uint) *HashScheduler {
	hs := &HashScheduler{
		chDie:    make(chan struct{}),
		chExit:   make(chan struct{}),
		chTasks:  make(chan scheduler.Task, 1<<8),
		started:  0,
		closed:   0,
		schedTag: tag,
		index:    index,
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("panic: %v\n%s", err.(error), string(debug.Stack()))
			}
		}()

		hs.digest()
	}()

	return hs
}

func (hs *HashScheduler) digest() {
	if atomic.AddInt32(&hs.started, 1) != 1 {
		return
	}

	defer func() {
		close(hs.chExit)
	}()

	for {
		select {
		case f := <-hs.chTasks:
			func() {
				defer func() {
					if err := recover(); err != nil {
						log.Errorf("panic: %v\n%s", err.(error), string(debug.Stack()))
					}
				}()

				f()
			}()

		case <-hs.chDie:
			return
		}
	}
}

// Close closes scheduler
func (hs *HashScheduler) Close() {
	if atomic.AddInt32(&hs.closed, 1) != 1 {
		return
	}
	close(hs.chDie)
	<-hs.chExit
}

// Schedule implements scheduler.LocalSchduler.Schedule
func (hs *HashScheduler) Schedule(
	_ *session.Session, _ interface{}, task scheduler.Task) {
	hs.chTasks <- task
}
