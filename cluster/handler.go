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
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"reflect"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aura-studio/nano/cluster/clusterpb"
	"github.com/aura-studio/nano/component"
	"github.com/aura-studio/nano/env"
	"github.com/aura-studio/nano/log"
	"github.com/aura-studio/nano/message"
	"github.com/aura-studio/nano/packet"
	"github.com/aura-studio/nano/pipeline"
	"github.com/aura-studio/nano/scheduler"
	"github.com/aura-studio/nano/session"
)

type rpcHandler func(session *session.Session, msg *message.Message, noCopy bool)

// LocalHandler stores local handlers & serivces info
type LocalHandler struct {
	localServices map[string]*component.Service // all registered service
	localHandlers map[string]*component.Handler // all handler method

	mu             sync.RWMutex
	remoteServices map[string]map[string][]*clusterpb.MemberInfo
	versionDict    map[uint32]string

	pipeline    pipeline.Pipeline
	currentNode *Node
}

// newHandler creates a new LocalHandler
func newHandler(currentNode *Node) *LocalHandler {
	h := &LocalHandler{
		localServices:  make(map[string]*component.Service),
		localHandlers:  make(map[string]*component.Handler),
		remoteServices: map[string]map[string][]*clusterpb.MemberInfo{},
		versionDict:    map[uint32]string{},
		pipeline:       currentNode.Pipeline,
		currentNode:    currentNode,
	}

	return h
}

// NewHandler creates a new handler without node
func NewHandler() *LocalHandler {
	h := &LocalHandler{
		localServices:  make(map[string]*component.Service),
		localHandlers:  make(map[string]*component.Handler),
		remoteServices: map[string]map[string][]*clusterpb.MemberInfo{},
	}

	return h
}

// Register register component on LocalHandler
func (h *LocalHandler) Register(comp component.Component, opts []component.Option) error {
	s := component.NewService(comp, opts)

	if _, ok := h.localServices[s.Name]; ok {
		return fmt.Errorf("handler: service already defined: %s", s.Name)
	}

	if err := s.ExtractHandler(); err != nil {
		return err
	}

	// register all localHandlers
	h.localServices[s.Name] = s
	for name, handler := range s.Handlers {
		route := fmt.Sprintf("%s.%s", s.Name, name)
		h.localHandlers[route] = handler
		message.WriteDictionaryItem(route, handler.Code)
	}

	return nil
}

func (h *LocalHandler) initMembers(members []*clusterpb.MemberInfo) {
	for _, m := range members {
		h.addMember(m)
	}
}

func (h *LocalHandler) addMember(member *clusterpb.MemberInfo) {
	h.mu.Lock()
	defer h.mu.Unlock()

	v := member.Version
	l := member.Label
	for _, s := range member.Services {
		if member.Version != "" {
			log.Infof("Register remote service %s(%s) from %s", s, v, l)
		} else {
			log.Infof("Register remote service %s from %s", s, l)
		}
		if _, ok := h.remoteServices[s]; !ok {
			h.remoteServices[s] = make(map[string][]*clusterpb.MemberInfo)
		}
		h.remoteServices[s][v] = append(h.remoteServices[s][v], member)
		h.versionDict[message.ShortVersion(v)] = v
	}

	message.WriteDictionary(member.Dictionary)
}

func (h *LocalHandler) delMember(addr string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for s, versionServices := range h.remoteServices {
		for v, members := range versionServices {
			for i, member := range members {
				l := member.Label
				if addr == member.ServiceAddr {
					if i == len(members)-1 {
						members = members[:i]
					} else {
						members = append(members[:i], members[i+1:]...)
					}

					if member.Version != "" {
						log.Infof("Deregister remote service %s(%s) from %s", s, v, l)
					} else {
						log.Infof("Deregister remote service %s from %s", s, l)
					}
					break
				}
			}
			if len(members) == 0 {
				delete(h.remoteServices[s], v)
			} else {
				h.remoteServices[s][v] = members
			}
		}
		if len(h.remoteServices[s]) == 0 {
			delete(h.remoteServices, s)
		}
	}
}

// FindVersions finds all versions for one service in cluster
func (h *LocalHandler) FindVersions(service string) []string {
	var versions []string

	// Only lock remote services when read services
	h.mu.RLock()
	if s, ok := h.remoteServices[service]; ok {
		for version := range s {
			var found bool
			for _, v := range versions {
				if version == v {
					found = true
					break
				}
			}
			if !found {
				versions = append(versions, version)
			}
		}
	}
	h.mu.RUnlock()

	if _, ok := h.localServices[service]; ok {
		version := env.Version
		var found bool
		for _, v := range versions {
			if version == v {
				found = true
				break
			}
		}
		if !found {
			versions = append(versions, version)
		}
	}

	sort.Strings(versions)
	return versions
}

// LocalService transforms local services info from map to slice
func (h *LocalHandler) LocalService() []string {
	var result []string
	for service := range h.localServices {
		result = append(result, service)
	}
	sort.Strings(result)
	return result
}

// RemoteService transforms remote services info from map to slice
func (h *LocalHandler) RemoteService() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var result []string
	for service := range h.remoteServices {
		result = append(result, service)
	}
	sort.Strings(result)
	return result
}

// LocalDictionary transforms local services info from map to slice
func (h *LocalHandler) LocalDictionary() []*clusterpb.DictionaryItem {
	var result []*clusterpb.DictionaryItem
	for name, handler := range h.localHandlers {
		result = append(result, &clusterpb.DictionaryItem{
			Route: name,
			Code:  uint32(handler.Code),
			Type:  handler.Type.String(),
		})
	}
	return result
}

// RouteHandler routes handler from localHandlers by route
func (h *LocalHandler) RouteHandler(route string) (*component.Handler, error) {
	handler, found := h.localHandlers[route]
	if !found {
		return nil, fmt.Errorf("Handler is not found by route")
	}
	return handler, nil
}

func (h *LocalHandler) handle(conn net.Conn) {
	// create a client agent and startup write gorontine
	agent := newAgent(conn, h.pipeline, h.processMessage, h.currentNode.Codec, h.currentNode.NodeID)
	h.currentNode.storeSession(agent.session)

	session.Created(agent.session)

	// startup write goroutine
	go agent.write()

	if env.Debug {
		log.Infof("New session established: %s", agent.String())
	}

	// guarantee agent related resource be destroyed
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("Handle session panic: %+v\n%s", err, debug.Stack())
		}
		request := &clusterpb.SessionClosedRequest{
			SessionID: agent.session.ID(),
		}

		h.currentNode.deleteSession(agent.session)

		for _, remote := range agent.session.RemoteSessionAddrs() {
			pool, err := h.currentNode.rpcClient.getConnPool(remote)
			if err != nil {
				log.Errorln("Cannot retrieve connection pool for address", remote, err)
				continue
			}
			client := clusterpb.NewMemberClient(pool.Get())
			_, err = client.SessionClosed(context.Background(), request)
			if err != nil {
				log.Errorln("Cannot closed session in remote address", remote, err)
				continue
			}
			if env.Debug {
				log.Infoln("Session closed notify remote server success", remote)
			}
		}

		agent.Close()

		session.Closed(agent.session)

		if env.Debug {
			log.Infof("Session read goroutine exit, NID=%d, SID=%d, UID=%d", agent.session.NID(), agent.session.SID(), agent.session.UID())
		}
	}()

	// read loop
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Infof("Read [%s], session will be closed immediately", err.Error())
			}
			return
		}

		if agent.payloadLength < 1024 {
			agent.payloadLength += n
		}

		packets, err := agent.codecEntity.DecodePacket(buf[:n])
		if err != nil {
			log.Errorln(err.Error())
			return
		}

		if len(packets) < 1 {
			continue
		}

		// process all packet
		for i := range packets {
			if err := h.processPacket(agent, packets[i]); err != nil {
				log.Errorln(err.Error())
				return
			}
		}
	}
}

func (h *LocalHandler) processPacket(agent *agent, p *packet.Packet) error {
	if !env.Debug {
		// now := time.Now().Unix()
		// if agent.packetSpeedLimitTimestamp == now {
		// 	agent.packetSpeedLimitCount++
		// 	if agent.packetSpeedLimitCount > 10 {
		// 		return fmt.Errorf("Too many packets received in a second")
		// 	}
		// } else {
		// 	agent.packetSpeedLimitTimestamp = now
		// 	agent.packetSpeedLimitCount = 1
		// }
	}

	msg, err := agent.codecEntity.DecodeMessage(p.Data)
	if err != nil {
		return err
	}

	if !agent.session.VersionBound {
		h.mu.RLock()
		version := h.versionDict[msg.ShortVer]
		h.mu.RUnlock()
		agent.session.BindShortVer(msg.ShortVer)
		agent.session.BindVersion(version)
		agent.session.BindBranch(msg.Branch)
		agent.session.VersionBound = true

		session.Inited(agent.session)
	}

	agent.session.DataType.Store(msg.DataType)
	agent.session.CryptoType.Store(msg.CryptoType)

	// Check message id
	maxMid := agent.session.MaxMid.Load()
	if msg.ID != maxMid+1 {
		return fmt.Errorf("invalid message id: remote: %d, local: %d, delta: %d", msg.ID, maxMid+1, msg.ID-maxMid-1)
	}
	agent.session.MaxMid.Store(msg.ID)

	// Check message time is not safe
	// now := time.Now().Unix()
	// if msg.ID > 1 && int64(msg.UnixTime) > now+10 || int64(msg.UnixTime) < now-10 {
	// 	return fmt.Errorf("invalid message time: remote: %d, local: %d, cost: %d(s)", msg.UnixTime, now, now-int64(msg.UnixTime))
	// }

	// Check message session id
	if agent.codecEntity.String() == "cryptocodec" {
		if msg.ID > 1 && msg.SessionID != agent.session.ID() {
			return fmt.Errorf("invalid message session id: remote: %d, local: %d", msg.SessionID, agent.session.ID())
		}
	}

	h.processMessage(agent.session, msg, false)

	agent.lastAt = time.Now().Unix()
	return nil
}

func (h *LocalHandler) findMembers(service string, shortVer uint32) (string, []*clusterpb.MemberInfo) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	version := h.versionDict[shortVer]
	if len(h.remoteServices[service][version]) > 0 {
		return version, h.remoteServices[service][version]
	}
	return version, h.remoteServices[service][""]
}

func (h *LocalHandler) remoteProcess(s *session.Session, msg *message.Message, noCopy bool) {
	index := strings.LastIndex(msg.Route, ".")
	if index < 0 {
		log.Errorf("nano/handler: invalid route %s", msg.Route)
		return
	}

	service := msg.Route[:index]
	var (
		remoteAddr string
	)
	if h.currentNode.Etcd {
		remoteAddr = fmt.Sprintf("etcd:///%s/%s", service, s.Version())

		if env.Debug {
			log.Infof("Type=%s, Route=%s, NID=%d, SID=%d, Version=%s, Branch=%d, UID=%d, MID=%d, Data=%dbytes",
				msg.Type.String(), msg.Route, s.NID(), s.SID(), s.Version(), s.Branch(), s.UID(), msg.ID, len(msg.Data))
		}
	} else {
		version, members := h.findMembers(service, msg.ShortVer)
		if len(members) == 0 {
			log.Errorf("nano/handler: %s (version:%s) not found(forgot registered?)", msg.Route, version)
			return
		}

		if env.Debug {
			log.Infof("Type=%s, Route=%s, NID=%d, SID=%d, Version=%s, Branch=%d, UID=%d, MID=%d, Data=%dbytes",
				msg.Type.String(), msg.Route, s.NID(), s.SID(), s.Version(), s.Branch(), s.UID(), msg.ID, len(msg.Data))
		}

		// Select a remote service address
		// 1. Use the service address directly if the router contains binding item
		// 2. Select a remote service address randomly and bind to router
		if addr, found := s.Router().Find(service); found {
			remoteAddr = addr
		} else {
			remoteAddr = members[rand.Intn(len(members))].ServiceAddr
			s.Router().Bind(service, remoteAddr)
		}
	}
	pool, err := h.currentNode.rpcClient.getConnPool(remoteAddr)
	if err != nil {
		log.Errorln(err)
		return
	}
	var data = msg.Data
	if !noCopy && len(msg.Data) > 0 {
		data = make([]byte, len(msg.Data))
		copy(data, msg.Data)
	}

	// Retrieve gate address and session ID
	gateAddr := h.currentNode.MemberAddr
	sessionID := s.ID()
	switch v := s.NetworkEntity().(type) {
	case *acceptor:
		gateAddr = v.gateAddr
		sessionID = v.sid
	}

	client := clusterpb.NewMemberClient(pool.Get())
	switch msg.Type {
	case message.Request:
		request := &clusterpb.RequestMessage{
			GateAddr:  gateAddr,
			SessionID: sessionID,
			ShortVer:  s.ShortVer(),
			ID:        msg.ID,
			UID:       s.UID(),
			Route:     msg.Route,
			DataType:  msg.DataType,
			Data:      data,
			RemoteAddr: &clusterpb.NetAddr{
				Network: s.RemoteAddr().Network(),
				Addr:    s.RemoteAddr().String(),
			},
			Branch:     s.Branch(),
			CryptoType: msg.CryptoType,
		}
		_, err = client.HandleRequest(context.Background(), request)
	case message.Notify:
		request := &clusterpb.NotifyMessage{
			GateAddr:  gateAddr,
			SessionID: sessionID,
			ShortVer:  s.ShortVer(),
			ID:        msg.ID,
			UID:       s.UID(),
			Route:     msg.Route,
			DataType:  msg.DataType,
			Data:      data,
			RemoteAddr: &clusterpb.NetAddr{
				Network: s.RemoteAddr().Network(),
				Addr:    s.RemoteAddr().String(),
			},
			Branch:     s.Branch(),
			CryptoType: msg.CryptoType,
		}
		_, err = client.HandleNotify(context.Background(), request)
	}
	if err != nil {
		log.Errorf("Process remote message (%d:%s) error: %+v",
			msg.ID, msg.Route, err)
	}
}

func (h *LocalHandler) processMessage(s *session.Session, msg *message.Message, noCopy bool) {
	var lastMid uint64
	switch msg.Type {
	case message.Request:
		lastMid = msg.ID
	case message.Notify:
		lastMid = msg.ID
	default:
		log.Errorln("Invalid message type: " + msg.Type.String())
		return
	}

	handler, found := h.localHandlers[msg.Route]
	if !found {
		h.remoteProcess(s, msg, noCopy)
	} else {
		h.localProcess(handler, lastMid, s, msg)
	}
}

func (h *LocalHandler) localProcess(handler *component.Handler, lastMid uint64, s *session.Session, msg *message.Message) {
	s = session.Context(s, lastMid)
	if pipe := h.pipeline; pipe != nil {
		err := pipe.Inbound().Process(s, msg)
		if err != nil {
			log.Errorln("Pipeline process failed: " + err.Error())
			return
		}
	}

	var payload = msg.Data
	var data interface{}
	if handler.IsRawArg {
		data = payload
	} else {
		data = reflect.New(handler.Type.Elem()).Interface()

		err := s.Deserialize(payload, data)
		if err != nil {
			log.Errorf("Deserialize to %T failed: %+v (%v)", data, err, payload)
			return
		}
	}

	if env.Debug {
		switch d := data.(type) {
		case []byte:
			log.Infof("Type=%s, Route=%s, NID=%d, SID=%d, Version=%s, Branch=%d, UID=%d, MID=%d, Data=%dbytes",
				msg.Type.String(), msg.Route, s.NID(), s.SID(), s.Version(), s.Branch(), s.UID(), s.LastMid(), len(d))
		default:
			log.Infof("Type=%s, Route=%s, NID=%d, SID=%d, Version=%s, Branch=%d, UID=%d, MID=%d, Data=%+v",
				msg.Type.String(), msg.Route, s.NID(), s.SID(), s.Version(), s.Branch(), s.UID(), s.LastMid(), data)
		}
	}

	args := []reflect.Value{handler.Receiver, reflect.ValueOf(s), reflect.ValueOf(data)}

	task := func() {
		result := handler.Method.Func.Call(args)
		if len(result) > 0 {
			if err := result[0].Interface(); err != nil {
				log.Errorf("nano/handler: handler %s error: %+v", msg.Route, err)
			}
		}
	}

	index := strings.LastIndex(msg.Route, ".")
	if index < 0 {
		log.Errorf("nano/handler: invalid route %s", msg.Route)
		return
	}

	// A message can be dispatch to global thread or a user customized thread
	serviceName := msg.Route[:index]
	handlerName := msg.Route[index+1:]
	service, found := h.localServices[serviceName]
	if !found {
		log.Errorf("Service not found: %+v", serviceName)
	}

	context := &scheduler.Context{
		ServiceName: serviceName,
		HandlerName: handlerName,
		Data:        data,
	}

	service.ScheduleFunc(s, context, task)
}
