package io

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/aura-studio/nano"
	"github.com/aura-studio/nano/benchmark/testdata"
	"github.com/aura-studio/nano/component"
	"github.com/aura-studio/nano/connector"
	"github.com/aura-studio/nano/serialize/protobuf"
)

const (
	addr    = ":12345" // local address
	conc    = 1000     // concurrent client count
	timeout = 5        // max test time
)

var (
	components component.Components
)

func client(t *testing.T) {
	c := connector.NewConnector()

	if err := c.Start(addr); err != nil {
		panic(err)
	}
	c.On("pong", func(data interface{}) {
		// t.Log("pong received")
	})
	<-c.Ready()
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
	<-nano.Ready()
	for i := 0; i < conc; i++ {
		go client(t)
	}

	log.SetFlags(log.LstdFlags | log.Llongfile)

	sg := make(chan os.Signal)
	signal.Notify(sg, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL)

	select {
	case <-time.After(timeout * time.Second):
	case <-sg:
	}
}
