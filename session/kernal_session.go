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
type KernalSession struct {
	sync.RWMutex                                 // protect data
	id           int64                           // session global unique id
	VersionBound bool                            // session version bound
	shortVer     uint32                          // session short version
	version      string                          // session version
	uid          int64                           // binding user id
	entity       NetworkEntity                   // low-level network entity
	data         map[string]interface{}          // session data store
	router       *Router                         // store remote addr
	onEvents     map[interface{}][]EventCallback // call EventCallback after event trigged
}

// New returns a new session instance
// a NetworkEntity is a low-level network instance
func NewKernalSession(entity NetworkEntity, id int64) *KernalSession {
	return &KernalSession{
		id:       id,
		entity:   entity,
		data:     make(map[string]interface{}),
		router:   newRouter(),
		onEvents: make(map[interface{}][]EventCallback),
	}
}

// NetworkEntity returns the low-level network agent object
func (s *KernalSession) NetworkEntity() NetworkEntity {
	return s.entity
}

// Router returns the service router
func (s *KernalSession) Router() *Router {
	return s.router
}

// RPC sends message to remote server
func (s *KernalSession) RPC(mid uint64, route string, v interface{}) error {
	return s.entity.RPC(mid, route, v)
}

// Push message to client
func (s *KernalSession) Push(route string, v interface{}) error {
	return s.entity.Push(route, v)
}

// Response message to client
func (s *KernalSession) Response(mid uint64, route string, v interface{}) error {
	return s.entity.Response(mid, route, v)
}
