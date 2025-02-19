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
	"github.com/aura-studio/nano/packet"
	"github.com/aura-studio/nano/pipeline"
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
		state    int32               // current agent state
		chDie    chan struct{}       // wait for close
		chSend   chan pendingMessage // push message queue
		lastAt   int64               // last heartbeat unix time stamp
		pipeline pipeline.Pipeline

		rpcHandler rpcHandler
		srv        reflect.Value // cached session reflect.Value

		codecEntity   codec.CodecEntity
		payloadLength int

		packetSpeedLimitTimestamp int64
		packetSpeedLimitCount     int32
	}

	pendingMessage struct {
		typ     message.Type // message type
		route   string       // message route(push)
		mid     uint64       // response message id(response)
		payload interface{}  // payload
	}
)

// Create new agent instance
func newAgent(conn net.Conn, pipeline pipeline.Pipeline, rpcHandler rpcHandler,
	codec codec.Codec, nodeID uint32) *agent {
	a := &agent{
		conn:        conn,
		state:       statusStart,
		chDie:       make(chan struct{}),
		lastAt:      time.Now().Unix(),
		chSend:      make(chan pendingMessage, agentWriteBacklog),
		pipeline:    pipeline,
		rpcHandler:  rpcHandler,
		codecEntity: codec.Entity(message.DuplicateDictionary()),
	}

	// binding session
	var sid uint64
	if env.SnowflakeNode != nil {
		sid = uint64(env.SnowflakeNode.Generate().Int64())
	} else {
		sid = uint64(nodeID)<<32 + uint64(service.Connections.SessionID())
	}

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
			log.Infof("Type=Push, Route=%s, NID=%d, SID=%d, Version=%s, Branch=%d, UID=%d, MID=%d, Data=%dbytes",
				route, a.session.NID(), a.session.SID(), a.session.Version(), a.session.Branch(), a.session.UID(), 0, len(d))
		default:
			log.Infof("Type=Push, Route=%s, NID=%d, SID=%d, Version=%s, Branch=%d, UID=%d, MID=%d, Data=%+v",
				route, a.session.NID(), a.session.SID(), a.session.Version(), a.session.Branch(), a.session.UID(), 0, v)
		}
	}

	return a.send(pendingMessage{typ: message.Push, route: route, payload: v})
}

func (a *agent) RPC(mid uint64, route string, v interface{}) error {
	if a.session.UID() > 0 {
		if a.status() == statusClosed {
			return ErrBrokenPipe
		}
	}

	if env.Debug {
		switch d := v.(type) {
		case []byte:
			log.Infof("Type=Notify, Route=%s, NID=%d, SID=%d, Version=%s, Branch=%d, UID=%d, MID=%d, Data=%dbytes",
				route, a.session.NID(), a.session.SID(), a.session.Version(), a.session.Branch(), a.session.UID(), mid, len(d))
		default:
			log.Infof("Type=Notify, Route=%s, NID=%d, SID=%d, Version=%s, Branch=%d, UID=%d, MID=%d, Data=%+v",
				route, a.session.NID(), a.session.SID(), a.session.Version(), a.session.Branch(), a.session.UID(), mid, v)
		}
	}

	data, err := a.session.Serialize(v)
	if err != nil {
		return err
	}

	msg := &message.Message{
		Type:       message.Notify,
		Branch:     a.session.Branch(),
		ShortVer:   a.session.ShortVer(),
		ID:         mid,
		Route:      route,
		DataType:   a.session.DataType.Load(),
		Data:       data,
		CryptoType: a.session.CryptoType.Load(),
	}
	a.rpcHandler(a.session, msg, true)
	return nil
}

// ResponseMid, implementation for session.NetworkEntity interface
// Response message to session
func (a *agent) Response(mid uint64, route string, v interface{}) error {
	if a.status() == statusClosed {
		return ErrBrokenPipe
	}

	if len(a.chSend) >= agentWriteBacklog {
		return ErrBufferExceed
	}

	if env.Debug {
		switch d := v.(type) {
		case []byte:
			log.Infof("Type=Response, Route=%s, NID=%d, SID=%d, Version=%s, Branch=%d, UID=%d, MID=%d, Data=%dbytes",
				route, a.session.NID(), a.session.SID(), a.session.Version(), a.session.Branch(), a.session.UID(), mid, len(d))
		default:
			log.Infof("Type=Response, Route=%s, NID=%d, SID=%d, Version=%s, Branch=%d, UID=%d, MID=%d, Data=%+v",
				route, a.session.NID(), a.session.SID(), a.session.Version(), a.session.Branch(), a.session.UID(), mid, v)
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
		log.Infof("Session closed, NID=%d, SID=%d, UID=%d, IP=%s",
			a.session.NID(), a.session.SID(), a.session.UID(), a.conn.RemoteAddr())
	}

	// prevent closing closed channel
	select {
	case <-a.chDie:
		// expect
	default:
		close(a.chDie)
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
			log.Infof("Session write goroutine exit, NID=%d, SID=%d, UID=%d",
				a.session.NID(), a.session.SID(), a.session.UID())
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

		case pendingMsg := <-a.chSend:
			data, err := a.session.Serialize(pendingMsg.payload)
			if err != nil {
				log.Errorln(err.Error())
				break
			}

			// construct message and encode
			msg := &message.Message{
				Type:       pendingMsg.typ,
				Branch:     a.session.Branch(),
				ShortVer:   a.session.ShortVer(),
				ID:         pendingMsg.mid,
				UnixTime:   uint32(time.Now().Unix()),
				SessionID:  a.session.ID(),
				Route:      pendingMsg.route,
				DataType:   a.session.DataType.Load(),
				Data:       data,
				CryptoType: a.session.CryptoType.Load(),
			}

			if pipe := a.pipeline; pipe != nil {
				err := pipe.Outbound().Process(a.session, msg)
				if err != nil {
					log.Errorln("broken pipeline", err.Error())
					break
				}
			}

			em, err := a.codecEntity.EncodeMessage(msg)
			if err != nil {
				log.Errorln(err.Error())
				break
			}

			// packet encode
			packets := []*packet.Packet{{Length: len(em), Data: em}}
			p, err := a.codecEntity.EncodePacket(packets)
			if err != nil {
				log.Errorln(err)
				break
			}

			chWrite <- p

		case <-a.chDie: // agent closed signal
			return

		case <-env.ConnDie: // application quit
			return
		}
	}
}
