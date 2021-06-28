package io

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/aura-studio/nano/benchmark/testdata"
	"github.com/aura-studio/nano/session"

	"github.com/aura-studio/nano/component"
)

//TestHandler is a component
type TestHandler struct {
	component.Base
	metrics int32
}

// AfterInit called after service init
func (h *TestHandler) AfterInit() {
	ticker := time.NewTicker(time.Second)

	// metrics output ticker
	go func() {
		for range ticker.C {
			qps := atomic.LoadInt32(&h.metrics)
			println("QPS", qps)
			if qps == 0 {
				panic(fmt.Errorf("QPS is 0"))
			}
			atomic.StoreInt32(&h.metrics, 0)
		}
	}()
}

// Ping is to push a Pong after received a Ping
func (h *TestHandler) Ping(s *session.Session, data *testdata.Ping) error {
	atomic.AddInt32(&h.metrics, 1)
	return s.Push("pong", &testdata.Pong{Content: data.Content})
}
