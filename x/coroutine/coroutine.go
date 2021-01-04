package coroutine

import (
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/lonng/nano/log"
)

type (
	//StatusType is an alias for string
	StatusType = string

	//ID is the unique identifier for coroutine
	ID = string

	// Coroutine is a simulator struct for coroutine
	Coroutine struct {
		co          ID
		status      StatusType
		inCh        chan []interface{}
		outCh       chan []interface{}
		fn          func(co ID, args ...interface{}) error
		mutexStatus *sync.Mutex
		mutexResume *sync.Mutex
	}
)

const (
	// Created means ID is created and not started.
	Created = "Created"

	// Suspended means ID is started and yielded.
	Suspended = "Suspended"

	// Running means ID is started and running.
	Running = "Running"

	// Dead means ID not created or ended.
	Dead = "Dead"
)

var (
	coroutines sync.Map
)

// Start wraps and starts a ID up.
// It is thread-safe, and it should be called before other funcs.
func Start(fn func(co ID) error) error {
	return Call(Wrap(func(co ID, args ...interface{}) error {
		return fn(co)
	}))
}

// Wrap wraps a ID and waits for a startup.
// It is thread-safe, and it should be called before other funcs.
// Call `Call` after `Wrap` to start up a ID.
func Wrap(fn func(co ID, args ...interface{}) error) ID {
	co := uuid.Must(uuid.NewV4()).String()
	entity := &Coroutine{
		co:          co,
		status:      Created,
		inCh:        make(chan []interface{}, 1),
		outCh:       make(chan []interface{}, 1),
		fn:          fn,
		mutexStatus: &sync.Mutex{},
		mutexResume: &sync.Mutex{},
	}
	coroutines.Store(co, entity)
	return co
}

// Call launch a ID that is already wrapped.
// It is not thread-safe, and it can only be called beside after Wrap.
// Call `Call` After `Wrap` to start up a ID.
func Call(co ID, args ...interface{}) error {
	e := findEntity(co)
	e.writeSyncStatus(Running)

	return func() (err error) {
		defer func() {
			if v := recover(); v != nil {
				log.Errorf("panic: %v\n%s", v.(error), string(debug.Stack()))
				err = fmt.Errorf("Coroutine ID %s called failed", co)
			}
		}()
		defer func() {
			coroutines.Delete(co)
		}()

		return e.fn(co, args...)
	}()
}

// Create wraps and yields a ID with no args, waits for a resume.
// It is not thread-safe, and it should be called before other funcs.
// Call `Resume` after `Create` to start up a ID.
func Create(fn func(co ID, inData ...interface{}) error) ID {
	co := Wrap(func(co ID, args ...interface{}) error {
		inData := Yield(co)
		return fn(co, inData...)
	})
	go Call(co)
	return co
}

// Resume continues a suspened ID, passing data in and out.
// It is thread-safe, and it can only be called in other Goroutine.
// Call `Resume` after `Create` to start up a ID.
// Call `Resume` after `Yield` to continue a ID.
func Resume(co ID, inData ...interface{}) ([]interface{}, bool) {
	e := findEntity(co)

	e.mutexResume.Lock()
	defer e.mutexResume.Unlock()
	if e.readSyncStatus() == Dead {
		return nil, false
	}
	outData := e.resume(inData)

	return outData, true
}

// TryResume likes Resume, but checks status instead of waiting for status.
// It is thread-safe, and it can only be called in other Goroutine.
// Call `TryResume` after `Create` to start up a ID.
// Call `TryResume` after `Yield` to continue a ID.
func TryResume(co ID, inData ...interface{}) ([]interface{}, bool) {
	e := findEntity(co)

	e.mutexResume.Lock()
	defer e.mutexResume.Unlock()

	if e.readSyncStatus() != Suspended {
		return nil, false
	}
	outData := e.resume(inData)

	return outData, true
}

// AsyncResume likes Resume, but works async.
// It is thread-safe, and it can only be called in other Goroutine.
// Call `AsyncResume` after `Create` to start up a ID.
// Call `AsyncResume` after `Yield` to continue a ID.
func AsyncResume(co ID, fn func(outData ...interface{}), inData ...interface{}) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("panic: %v\n%s", err.(error), string(debug.Stack()))
			}
		}()

		e := findEntity(co)
		e.mutexResume.Lock()
		defer e.mutexResume.Unlock()
		if e.readSyncStatus() == Dead {
			panic(fmt.Errorf("ID is Dead"))
		}
		outData := e.resume(inData)

		fn(outData...)
	}()
}

// Yield suspends a running coroutine, passing data in and out.
// It is not thread-safe, and it can only be called in entity.fn.
// Call `Resume`, `TryResume` or `AsyncResume`
// after `Yield` to continue a ID.
func Yield(co ID, outData ...interface{}) []interface{} {
	e := findEntity(co)
	e.writeSyncStatus(Suspended)
	inData := e.yield(outData)
	e.writeSyncStatus(Running)
	return inData
}

// Status shows the status of a ID.
// It is thread-safe, and it can be called in any Goroutine.
// Call `Status` anywhere you need.
func Status(co ID) StatusType {
	entity, ok := coroutines.Load(co)
	if !ok {
		return Dead
	}
	e := entity.(*Coroutine)
	return e.readSyncStatus()
}

func findEntity(co ID) *Coroutine {
	entity, ok := coroutines.Load(co)
	if !ok {
		panic(fmt.Errorf("ID %s is not found in entity map", co))
	}
	e := entity.(*Coroutine)
	return e
}

func (e *Coroutine) writeSyncStatus(status StatusType) {
	e.mutexStatus.Lock()
	defer e.mutexStatus.Unlock()
	e.status = status
}

func (e *Coroutine) readSyncStatus() StatusType {
	e.mutexStatus.Lock()
	defer e.mutexStatus.Unlock()
	return e.status
}

func (e *Coroutine) resume(inData []interface{}) []interface{} {
	var outData []interface{}

	select {
	case outData = <-e.outCh:
		break
	case <-time.After(time.Duration(expire) * time.Second):
		panic(fmt.Errorf("ID %s suspended for too long", e.co))
	}

	select {
	case e.inCh <- inData:
		break
	case <-time.After(time.Duration(expire) * time.Second):
		panic(fmt.Errorf("ID %s suspended for too long", e.co))
	}

	return outData
}

func (e *Coroutine) yield(outData []interface{}) []interface{} {
	var inData []interface{}

	select {
	case e.outCh <- outData:
		break
	case <-time.After(time.Duration(expire) * time.Second):
		e.writeSyncStatus(Dead)
		panic(fmt.Errorf("ID %s suspended for too long", e.co))
	}

	select {
	case inData = <-e.inCh:
		break
	case <-time.After(time.Duration(expire) * time.Second):
		e.writeSyncStatus(Dead)
		panic(fmt.Errorf("ID %s suspended for too long", e.co))
	}

	return inData
}
