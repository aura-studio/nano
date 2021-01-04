package sched

import (
	"github.com/lonng/nano/log"
	"github.com/lonng/nano/scheduler"
)

func WrapTask(v interface{}) scheduler.Task {
	switch task := v.(type) {
	case func():
		return task
	case func() error:
		return func() {
			if err := task(); err != nil {
				log.Errorf("Schedule task error: %+v", err)
			}
		}
	default:
		panic("Unsupported task type")
	}
}
