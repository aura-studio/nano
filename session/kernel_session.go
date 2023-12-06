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

package session

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/aura-studio/nano/serializer"
)

// NetworkEntity represent low-level network instance
type NetworkEntity interface {
	Push(route string, v interface{}) error
	RPC(mid uint64, route string, v interface{}) error
	Response(mid uint64, route string, v interface{}) error
	Close() error
	RemoteAddr() net.Addr
}

// EventCallback is the func called after event trigged
type EventCallback func(*Session, ...interface{})

// Session represents a client session which could storage temp data during low-level
// keep connected, all data will be released when the low-level connection was broken.
// Session instance related to the client will be passed to Handler method as the first
// parameter.
type KernelSession struct {
	sync.RWMutex                                       // protect data
	id                 uint64                          // session global unique id
	branch             atomic.Uint32                   // logic branch
	VersionBound       bool                            // session version bound
	shortVer           uint32                          // session short version
	version            string                          // session version
	uid                int64                           // binding user id
	entity             NetworkEntity                   // low-level network entity
	data               map[string]interface{}          // session data store
	router             *Router                         // store remote addr
	onEvents           map[interface{}][]EventCallback // call EventCallback after event trigged
	remoteSessionAddrs sync.Map                        // rpc addr
	MaxMid             atomic.Uint64
	DataType           atomic.Uint32
}

// New returns a new session instance
// a NetworkEntity is a low-level network instance
func NewKernelSession(entity NetworkEntity, id uint64) *KernelSession {
	return &KernelSession{
		id:       id,
		entity:   entity,
		data:     make(map[string]interface{}),
		router:   newRouter(),
		onEvents: make(map[interface{}][]EventCallback),
	}
}

// NetworkEntity returns the low-level network agent object
func (s *KernelSession) NetworkEntity() NetworkEntity {
	return s.entity
}

// Router returns the service router
func (s *KernelSession) Router() *Router {
	return s.router
}

// RPC sends message to remote server
func (s *KernelSession) RPC(mid uint64, route string, v interface{}) error {
	return s.entity.RPC(mid, route, v)
}

// Push message to client
func (s *KernelSession) Push(route string, v interface{}) error {
	return s.entity.Push(route, v)
}

// Response message to client
func (s *KernelSession) Response(mid uint64, route string, v interface{}) error {
	return s.entity.Response(mid, route, v)
}

func (s *KernelSession) AddRemoteSessionAddr(addr string) {
	s.remoteSessionAddrs.Store(addr, struct{}{})
}

func (s *KernelSession) RemoteSessionAddrs() []string {
	addrs := make([]string, 0)
	s.remoteSessionAddrs.Range(func(key, value interface{}) bool {
		addrs = append(addrs, key.(string))
		return true
	})
	return addrs
}

func (s *KernelSession) Serialize(v interface{}) ([]byte, error) {
	if data, ok := v.([]byte); ok {
		return data, nil
	}
	data, err := serializer.FromType(s.DataType.Load()).Marshal(v)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *KernelSession) Deserialize(data []byte, v interface{}) error {
	return serializer.FromType(s.DataType.Load()).Unmarshal(data, v)
}

// ID returns the session id
func (s *KernelSession) ID() uint64 {
	return s.id
}

// UID returns uid that bind to current session
func (s *KernelSession) UID() int64 {
	return atomic.LoadInt64(&s.uid)
}
