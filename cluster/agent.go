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

package cluster

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/aura-studio/nano/codec"
	"github.com/aura-studio/nano/env"
	"github.com/aura-studio/nano/log"
	"github.com/aura-studio/nano/message"
	"github.com/aura-studio/nano/pipeline"
	"github.com/aura-studio/nano/serialize"
	"github.com/aura-studio/nano/service"
	"github.com/aura-studio/nano/session"
)

const (
	agentWriteBacklog = 256
)

var (
	// ErrBrokenPipe represents the low-level connection has broken.
	ErrBrokenPipe = errors.New("broken low-level pipe")
	// ErrBufferExceed indicates that the current session buffer is full and
	// can not receive more data.
	ErrBufferExceed = errors.New("session send buffer exceed")
)

type (
	// Agent corresponding a user, used for store raw conn information
	agent struct {
		// regular agent member
		session  *session.Session    // session
		conn     net.Conn            // low-level conn fd
		lastMid  uint64              // last message id
		state    int32               // current agent state
		chDie    chan struct{}       // wait for close
		chSend   chan pendingMessage // push message queue
		lastAt   int64               // last heartbeat unix time stamp
		decoder  *codec.Decoder      // binary decoder
		pipeline pipeline.Pipeline

		rpcHandler  rpcHandler
		srv         reflect.Value                   // cached session reflect.Value
		routes      map[string]uint16               // copy system routes for agent
		codes       map[uint16]string               // copy system codes for agent
		serializers map[string]serialize.Serializer // copy system serializers for agent
		compressed  bool                            // whether to use compressed msg to client
		recvPckCnt  int64                           // agent receive packet count
		sendPckCnt  int64                           // agent send packet count
	}

	pendingMessage struct {
		typ     message.Type // message type
		route   string       // message route(push)
		mid     uint64       // response message id(response)
		payload interface{}  // payload
	}
)

// Create new agent instance
func newAgent(conn net.Conn, pipeline pipeline.Pipeline, rpcHandler rpcHandler) *agent {
	routes, codes := message.ReadDictionary()
	serializers := message.ReadSerializers()
	a := &agent{
		conn:        conn,
		state:       statusStart,
		chDie:       make(chan struct{}),
		lastAt:      time.Now().Unix(),
		chSend:      make(chan pendingMessage, agentWriteBacklog),
		decoder:     codec.NewDecoder(),
		pipeline:    pipeline,
		rpcHandler:  rpcHandler,
		routes:      routes,
		codes:       codes,
		serializers: serializers,
	}

	// binding session
	sid := service.Connections.SessionID()
	s := session.New(a, sid)
	a.session = s
	a.srv = reflect.ValueOf(s)

	return a
}

func (a *agent) send(m pendingMessage) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = ErrBrokenPipe
		}
	}()
	a.chSend <- m
	return
}

// LastMid implements the session.NetworkEntity interface
func (a *agent) LastMid() uint64 {
	return a.lastMid
}

// Push, implementation for session.NetworkEntity interface
func (a *agent) Push(route string, v interface{}) error {
	if a.status() == statusClosed {
		return ErrBrokenPipe
	}

	if len(a.chSend) >= agentWriteBacklog {
		return ErrBufferExceed
	}

	if env.Debug {
		switch d := v.(type) {
		case []byte:
			log.Infof("Type=Push, Route=%s, ID=%d, Version=%s, UID=%d,  MID=%d, Data=%dbytes",
				route, a.session.ID(), a.session.Version(), a.session.UID(), 0, len(d))
		default:
			log.Infof("Type=Push, Route=%s, ID=%d, Version=%s, UID=%d, MID=%d, Data=%+v",
				route, a.session.ID(), a.session.Version(), a.session.UID(), 0, v)
		}
	}

	return a.send(pendingMessage{typ: message.Push, route: route, payload: v})
}

// RPC, implementation for session.NetworkEntity interface
func (a *agent) RPC(route string, v interface{}) error {
	if a.status() == statusClosed {
		return ErrBrokenPipe
	}

	data, err := message.RouteSerialize(a.serializers, route, v)
	if err != nil {
		return err
	}

	if env.Debug {
		switch d := v.(type) {
		case []byte:
			log.Infof("Type=Notify, Route=%s, ID=%d, Version=%s, UID=%d, MID=%d, Data=%dbytes",
				route, a.session.ID(), a.session.Version(), a.session.UID(), a.lastMid, len(d))
		default:
			log.Infof("Type=Notify, Route=%s, ID=%d, Version=%s, UID=%d, MID=%d, Data=%+v",
				route, a.session.ID(), a.session.Version(), a.session.UID(), a.lastMid, v)
		}
	}

	msg := &message.Message{
		Type:     message.Notify,
		ShortVer: a.session.ShortVer(),
		ID:       a.lastMid,
		Route:    route,
		Data:     data,
	}
	a.rpcHandler(a.session, msg, true)
	return nil
}

// Response, implementation for session.NetworkEntity interface
// Response message to session
func (a *agent) Response(route string, v interface{}) error {
	return a.ResponseMid(a.lastMid, route, v)
}

// ResponseMid, implementation for session.NetworkEntity interface
// Response message to session
func (a *agent) ResponseMid(mid uint64, route string, v interface{}) error {
	if a.status() == statusClosed {
		return ErrBrokenPipe
	}

	if len(a.chSend) >= agentWriteBacklog {
		return ErrBufferExceed
	}

	if env.Debug {
		switch d := v.(type) {
		case []byte:
			log.Infof("Type=Response, Route=%s, ID=%d, Version=%s, UID=%d, MID=%d, Data=%dbytes",
				route, a.session.ID(), a.session.Version(), a.session.UID(), mid, len(d))
		default:
			log.Infof("Type=Response, Route=%s, ID=%d, Version=%s, UID=%d, MID=%d, Data=%+v",
				route, a.session.ID(), a.session.Version(), a.session.UID(), mid, v)
		}
	}

	return a.send(pendingMessage{typ: message.Response, route: route, mid: mid, payload: v})
}

// Close, implementation for session.NetworkEntity interface
// Close closes the agent, clean inner state and close low-level connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (a *agent) Close() error {
	if a.status() == statusClosed {
		return ErrCloseClosedSession
	}
	a.setStatus(statusClosed)

	if env.Debug {
		log.Infof("Session closed, ID=%d, UID=%d, IP=%s",
			a.session.ID(), a.session.UID(), a.conn.RemoteAddr())
	}

	// prevent closing closed channel
	select {
	case <-a.chDie:
		// expect
	default:
		close(a.chDie)
		session.Closed(a.session)
	}

	return a.conn.Close()
}

// RemoteAddr, implementation for session.NetworkEntity interface
// returns the remote network address.
func (a *agent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

// String, implementation for Stringer interface
func (a *agent) String() string {
	return fmt.Sprintf("Remote=%s, LastTime=%d", a.conn.RemoteAddr().String(),
		atomic.LoadInt64(&a.lastAt))
}

func (a *agent) status() int32 {
	return atomic.LoadInt32(&a.state)
}

func (a *agent) setStatus(state int32) {
	atomic.StoreInt32(&a.state, state)
}

func (a *agent) write() {
	chWrite := make(chan []byte, agentWriteBacklog)
	// clean func
	defer func() {
		close(a.chSend)
		close(chWrite)
		a.Close()
		if env.Debug {
			log.Infof("Session write goroutine exit, SessionID=%d, UID=%d",
				a.session.ID(), a.session.UID())
		}
	}()

	for {
		select {
		case data := <-chWrite:
			// close agent while low-level conn broken
			if _, err := a.conn.Write(data); err != nil {
				log.Errorln(err.Error())
				return
			}

		case data := <-a.chSend:
			payload, err := message.Serialize(data.payload)
			if err != nil {
				switch data.typ {
				case message.Push:
					log.Errorf("Push: %s error: %s", data.route, err.Error())
				case message.Response:
					log.Errorf("Response message(id: %d) error: %s", data.mid, err.Error())
				default:
					// expect
				}
				break
			}

			// construct message and encode
			m := &message.Message{
				Type:     data.typ,
				ShortVer: a.session.ShortVer(),
				Data:     payload,
				Route:    data.route,
				ID:       data.mid,
			}
			if pipe := a.pipeline; pipe != nil {
				err := pipe.Outbound().Process(a.session, m)
				if err != nil {
					log.Errorln("broken pipeline", err.Error())
					break
				}
			}

			var routes map[string]uint16
			if a.compressed {
				routes = a.routes
			}
			em, err := message.Encode(m, routes)
			if err != nil {
				log.Errorln(err.Error())
				break
			}

			// packet encode
			p, err := codec.Encode(em)
			if err != nil {
				log.Errorln(err)
				break
			}

			a.sendPckCnt++
			chWrite <- p

		case <-a.chDie: // agent closed signal
			return

		case <-env.Die: // application quit
			return
		}
	}
}
