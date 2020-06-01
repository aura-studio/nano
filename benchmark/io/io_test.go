package io

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/lonng/nano"
	"github.com/lonng/nano/benchmark/testdata"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/connector"
	"github.com/lonng/nano/serialize/protobuf"
)

const (
	addr     = ":12345" // local address
	conc     = 1000     // concurrent client count
	duration = 5        // max test time
	wait     = 1        // time to wait for server startup
)

var (
	components component.Components
)

func client(t *testing.T) {
	c := connector.NewConnector()

	chReady := make(chan struct{})
	c.OnConnected(func() {
		// t.Log("connected")
		chReady <- struct{}{}
	})

	if err := c.Start(addr); err != nil {
		panic(err)
	}

	c.On("pong", func(data interface{}) {
		// t.Log("pong received")
	})
	<-chReady
	for {
		err := c.Notify("TestHandler.Ping", &testdata.Ping{})
		if err != nil {
			t.Error(err)
		}
		// t.Log("ping send")
		time.Sleep(10 * time.Millisecond)
	}
}

func server(t *testing.T) {
	components.Register(&TestHandler{metrics: 0})
	nano.Listen(addr,
		nano.WithComponents(&components),
		nano.WithSerializer(protobuf.NewSerializer()),
	)
}

func TestPingPong(t *testing.T) {
	go server(t)

	// wait server startup
	time.Sleep(wait * time.Second)
	for i := 0; i < conc; i++ {
		go client(t)
	}

	log.SetFlags(log.LstdFlags | log.Llongfile)

	sg := make(chan os.Signal)
	signal.Notify(sg, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL)

	select {
	case <-time.After(duration * time.Second):
	case <-sg:
	}
}
