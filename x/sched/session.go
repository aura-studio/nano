package sched

import (
	"github.com/lonng/nano/env"
	"runtime/debug"
	"sync"
	"time"

	"github.com/lonng/nano/log"

	"github.com/lonng/nano/scheduler"
	"github.com/lonng/nano/session"

	"sync/atomic"
)

type (
	// SessionScheduler implements shceduler.LocalScheduler
	SessionScheduler struct {
		chDie   chan struct{}
		chExit  chan struct{}
		chTasks chan scheduler.Task
		started int32
		closed  int32
		session *session.Session
		timerManager *TimerManager
	}
)

var (
	// sessionSchedulers stores all *SessionScheduler
	sessionSchedulers sync.Map
)

// NewSessionScheduler returns a new SessionScheduler
func NewSessionScheduler(s *session.Session) *SessionScheduler {
	ss := &SessionScheduler{
		chDie:   make(chan struct{}),
		chExit:  make(chan struct{}),
		chTasks: make(chan scheduler.Task, 1<<8),
		started: 0,
		closed:  0,
		session: s,
		timerManager: &TimerManager{
			timers: map[int64]*Timer{},
			closingTimer: []int64 {},
			createdTimer: []*Timer {},
		},
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("panic: %v\n%s", err.(error), string(debug.Stack()))
			}
		}()

		ss.digest()
	}()

	sessionSchedulers.Store(s.ID(), ss)

	return ss
}

// RemoveSessionScheduler remove one SessionScheduler
func RemoveSessionScheduler(s *session.Session) {
	if v, ok := sessionSchedulers.Load(s.ID()); ok {
		v.(*SessionScheduler).Close()
		sessionSchedulers.Delete(s.ID())
	} else {
		log.Errorf("session scheduler not found,  SessionID=%v, UID=%v",
			s.ID(), s.UID())
	}
}

// InitSessionSchedulers use session scheduler for new session
// and removes it when session stopped.
func InitSessionSchedulers() {
	session.OnClosed(func(s *session.Session) {
		if s.Value("Scheduler") != nil {
			RemoveSessionScheduler(s)
			log.Infof("Session scheduler closed, SessionID=%v, UID=%v",
				s.ID(), s.UID())
		}
		s.Remove("Scheduler")
	})
}

// GetSessionScheduler get session scheduler, creating it when it's not created.
func GetSessionScheduler(s *session.Session) *SessionScheduler {
	if s.Value("Scheduler") != nil {
		return s.Value("Scheduler").(*SessionScheduler)
	}
	ss := NewSessionScheduler(s)
	s.Set("Scheduler", ss)
	log.Infof("Session scheduler inited, SessionID=%v, UID=%v",
		s.ID(), s.UID())
	return ss
}

// CloseSessionSchedulers closes all SessionScheduler
func CloseSessionSchedulers() {
	sessionSchedulers.Range(func(k, v interface{}) bool {
		v.(*SessionScheduler).Close()
		return true
	})
	sessionSchedulers = sync.Map{}
}

func (ss *SessionScheduler) digest() {
	if atomic.AddInt32(&ss.started, 1) != 1 {
		return
	}

	ticker := time.NewTicker(env.TimerPrecision)
	defer func() {
		ticker.Stop()
		close(ss.chExit)
	}()

	for {
		select {

		case <- ticker.C:
			cron(ss.timerManager)

		case f := <-ss.chTasks:
			func() {
				defer func() {
					if err := recover(); err != nil {
						log.Errorf("panic: %v\n%s", err.(error), string(debug.Stack()))
					}
				}()

				f()
			}()

		case <-ss.chDie:
			return
		}
	}
}

// Close closes scheduler
func (ss *SessionScheduler) Close() {
	if atomic.AddInt32(&ss.closed, 1) != 1 {
		return
	}
	close(ss.chDie)
	<-ss.chExit
}

// Schedule implements scheduler.Schedule
func (ss *SessionScheduler) Schedule(
	_ *session.Session, _ interface{}, task scheduler.Task) {
	ss.chTasks <- task
}
