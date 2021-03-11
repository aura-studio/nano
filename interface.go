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

package nano

import (
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/lonng/nano/cluster"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/env"
	"github.com/lonng/nano/log"
	"github.com/lonng/nano/scheduler"
)

const (
	Cluster   = "cluster"
	Singleton = "singleton"
	Frontend  = "frontend"
	Backend   = "backend"
)

var chReady = make(chan struct{}, 1)

var (
	// app represents the current server process
	app = &struct {
		running int32
		name    string    // current application name
		startAt time.Time // startup time
		mode    string    // cluster mode
		typ     string    // frontend or backend
	}{}
)

// Listen listens on the TCP network address addr
// and then calls Serve with handler to handle requests
// on incoming connections.
func Listen(addr string, opts ...Option) {
	if atomic.AddInt32(&app.running, 1) != 1 {
		log.Infoln("Nano server is running")
		return
	}

	// application initialize
	app.name = strings.TrimLeft(filepath.Base(os.Args[0]), "/")
	app.startAt = time.Now()

	// environment initialize
	if wd, err := os.Getwd(); err != nil {
		panic(err)
	} else {
		env.Wd, _ = filepath.Abs(wd)
	}

	opt := cluster.Options{
		Components: &component.Components{},
	}
	for _, option := range opts {
		option(&opt)
	}

	log.SetLogger(opt.Logger)

	log.Infoln("Nano server is starting...")

	// Use listen address as client address in non-cluster mode
	if !opt.IsMaster && opt.AdvertiseAddr == "" && opt.ClientAddr == "" {
		app.mode = Singleton
		opt.ClientAddr = addr
	} else {
		app.mode = Cluster
	}

	// Set the retry interval to 3 secondes if doesn't set by user
	if opt.RetryInterval == 0 {
		opt.RetryInterval = time.Second * 3
	}

	node := &cluster.Node{
		Options:     opt,
		ServiceAddr: addr,
	}

	err := node.Startup()
	if err != nil {
		log.Fatalf("Nano server startup failed: %v", err)
	}

	if node.ClientAddr != "" {
		app.typ = Frontend
	} else {
		app.typ = Backend
	}

	log.Infof("Startup %v as %v server in %v mode", app.name, app.typ, app.mode)

	if node.DebugAddr != "" {
		log.Infof("Debug address: %s", node.DebugAddr)
	}

	if node.ClientAddr != "" {
		if node.HttpUpgrader == nil {
			log.Infof("TCP address: %s", node.ClientAddr)
		} else if node.HttpAddr == "" {
			log.Infof("HTTP address: %s", node.ClientAddr)
		} else {
			log.Infof("TCP address: %s", node.ClientAddr)
			log.Infof("HTTP address: %s", node.HttpAddr)
		}
	}

	if node.ServiceAddr != node.ClientAddr {
		log.Infof("Service address: %s", node.ServiceAddr)
	}

	log.Infof("Nano server is serving...")

	go scheduler.Digest()
	sg := make(chan os.Signal, 1)
	signal.Notify(sg, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	chReady <- struct{}{}

	select {
	case <-env.Die:
		log.Infoln("The app will shutdown in a few seconds")
	case s := <-sg:
		log.Infoln("Nano server got signal", s)
	}

	log.Infoln("Nano server is stopping...")

	node.Shutdown()
	scheduler.Close()
	atomic.StoreInt32(&app.running, 0)
	log.Infoln("Nano server stopped")
}

// Shutdown send a signal to let 'nano' shutdown itself.
func Shutdown() {
	close(env.Die)
}

func Ready() <-chan struct{} {
	return chReady
}
