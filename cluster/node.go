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
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/lonng/nano/cluster/clusterpb"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/env"
	"github.com/lonng/nano/log"
	"github.com/lonng/nano/message"
	"github.com/lonng/nano/persistence"
	"github.com/lonng/nano/pipeline"
	"github.com/lonng/nano/session"
	"github.com/lonng/nano/upgrader"
	"google.golang.org/grpc"
)

// Options contains some configurations for current node
type Options struct {
	Pipeline       pipeline.Pipeline
	Convention     Convention
	MasterPersist  persistence.Persistence
	IsMaster       bool
	AdvertiseAddr  string
	RetryInterval  time.Duration
	ClientAddr     string
	DebugAddr      string
	Components     *component.Components
	Label          string
	HttpUpgrader   upgrader.Upgrader
	HttpAddr       string
	TSLCertificate string
	TSLKey         string
	Logger         log.Logger
}

// Node represents a node in nano cluster, which will contains a group of services.
// All services will register to cluster and messages will be forwarded to the node
// which provides respective service
type Node struct {
	Options            // current node options
	ServiceAddr string // current server service address (RPC)

	cluster      *cluster
	handler      *LocalHandler
	server       *grpc.Server
	rpcClient    *rpcClient
	conventioner *conventioner

	mu       sync.RWMutex
	sessions map[int64]*session.Session
}

// Startup bootstraps a start up.
func (n *Node) Startup() error {
	if n.ServiceAddr == "" {
		return errors.New("service address cannot be empty in master node")
	}
	n.sessions = map[int64]*session.Session{}
	n.cluster = newCluster(n)
	n.handler = newHandler(n)
	n.conventioner = newConventioner(n)
	components := n.Components.List()
	for _, c := range components {
		err := n.handler.Register(c.Comp, c.Opts)
		if err != nil {
			return err
		}
	}

	if err := n.initNode(); err != nil {
		return err
	}

	// Initialize all components
	for _, c := range components {
		c.Comp.Init()
	}
	for _, c := range components {
		c.Comp.AfterInit()
	}

	if n.DebugAddr != "" {
		go n.ListenAndServeDebug()
	}

	if n.ClientAddr != "" {
		if n.HttpUpgrader == nil {
			go n.listenAndServe()
		} else if n.HttpAddr == "" {
			go n.listenAndServeHttp()
		} else {
			go n.listenAndServe()
			go n.listenAndServeHttp()
		}
	}

	return nil
}

// Handler returns localhandler for this node.
func (n *Node) Handler() *LocalHandler {
	return n.handler
}

func (n *Node) initNode() error {
	// Current node is not master server and does not contains master
	// address, so running in singleton mode
	if !n.IsMaster && n.AdvertiseAddr == "" {
		return nil
	}

	listener, err := net.Listen("tcp", n.ServiceAddr)
	if err != nil {
		return err
	}

	// Initialize the gRPC server and register service
	n.server = grpc.NewServer()
	n.rpcClient = newRPCClient()
	clusterpb.RegisterMemberServer(n.server, n)

	go func() {
		err := n.server.Serve(listener)
		if err != nil {
			log.Fatalf("Start current node failed: %v", err)
		}
	}()

	if n.IsMaster {
		clusterpb.RegisterMasterServer(n.server, n.cluster)
		member := &Member{
			isMaster: true,
			memberInfo: &clusterpb.MemberInfo{
				Label:       n.Label,
				Version:     env.Version,
				ServiceAddr: n.ServiceAddr,
				Services:    n.handler.LocalService(),
				Dictionary:  n.handler.LocalDictionary(),
			},
		}
		n.cluster.members = append(n.cluster.members, member)
		n.cluster.setRPCClient(n.rpcClient)
		if n.MasterPersist != nil {
			var memberInfos []*clusterpb.MemberInfo
			if err := n.MasterPersist.Get(&memberInfos); err != nil {
				return err
			}
			for _, memberInfo := range memberInfos {
				n.cluster.members = append(n.cluster.members, &Member{isMaster: false, memberInfo: memberInfo})
				n.handler.addMember(memberInfo)
			}
		}
	} else {
		pool, err := n.rpcClient.getConnPool(n.AdvertiseAddr)
		if err != nil {
			return err
		}
		client := clusterpb.NewMasterClient(pool.Get())
		request := &clusterpb.RegisterRequest{
			MemberInfo: &clusterpb.MemberInfo{
				Label:       n.Label,
				Version:     env.Version,
				ServiceAddr: n.ServiceAddr,
				Services:    n.handler.LocalService(),
				Dictionary:  n.handler.LocalDictionary(),
			},
		}
		for {
			resp, err := client.Register(context.Background(), request)
			if err == nil {
				n.handler.initMembers(resp.Members)
				n.cluster.initMembers(resp.Members)
				break
			}
			log.Errorln("Register current node to cluster failed", err, "and will retry in", n.RetryInterval.String())
			time.Sleep(n.RetryInterval)
		}

	}

	return nil
}

// Shutdown all components registered by application, that
// call by reverse order against register
func (n *Node) Shutdown() {
	// reverse call `BeforeShutdown` hooks
	components := n.Components.List()
	length := len(components)
	for i := length - 1; i >= 0; i-- {
		components[i].Comp.BeforeShutdown()
	}

	// reverse call `Shutdown` hooks
	for i := length - 1; i >= 0; i-- {
		components[i].Comp.Shutdown()
	}

	if !n.IsMaster && n.AdvertiseAddr != "" {
		pool, err := n.rpcClient.getConnPool(n.AdvertiseAddr)
		if err != nil {
			log.Errorln("Retrieve master address error", err)
			goto EXIT
		}
		client := clusterpb.NewMasterClient(pool.Get())
		request := &clusterpb.UnregisterRequest{
			ServiceAddr: n.ServiceAddr,
		}
		_, err = client.Unregister(context.Background(), request)
		if err != nil {
			log.Errorln("Unregister current node failed", err)
			goto EXIT
		}
	}

EXIT:
	if n.server != nil {
		n.server.GracefulStop()
	}
}

// Enable current server accept connection
func (n *Node) listenAndServe() {
	listener, err := net.Listen("tcp", n.ClientAddr)
	if err != nil {
		log.Fatal(err.Error())
	}

	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorln(err.Error())
			continue
		}

		go n.handler.handle(conn)
	}
}

func (n *Node) listenAndServeHttp() {
	router := mux.NewRouter()
	router.HandleFunc("/{route:[A-Za-z\\.]*}", func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		conn, err := n.HttpUpgrader.Upgrade(w, r, params)
		if err != nil {
			log.Errorf("Upgrade failure, URI=%s, Error=%s", r.RequestURI, err.Error())
		}

		go n.handler.handle(conn)
	})
	http.Handle("/", router)

	var addr string
	if n.HttpAddr != "" {
		addr = n.HttpAddr
	} else {
		addr = n.ClientAddr
	}

	if len(n.TSLCertificate) != 0 {
		if err := http.ListenAndServeTLS(addr, n.TSLCertificate, n.TSLKey, nil); err != nil {
			log.Fatal(err.Error())
		}
	} else {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatal(err.Error())
		}
	}
}

func (n *Node) ListenAndServeDebug() {
	if err := http.ListenAndServe(n.DebugAddr, nil); err != nil {
		log.Fatal(err.Error())
	}
}

func (n *Node) storeSession(s *session.Session) {
	n.mu.Lock()
	n.sessions[s.ID()] = s
	n.mu.Unlock()
}

func (n *Node) findSession(sid int64) *session.Session {
	n.mu.RLock()
	s := n.sessions[sid]
	n.mu.RUnlock()
	return s
}

func (n *Node) deleteSession(s *session.Session) {
	n.mu.Lock()
	s, found := n.sessions[s.ID()]
	if found {
		delete(n.sessions, s.ID())
	}
	n.mu.Unlock()
}

func (n *Node) findOrCreateSession(sid int64, gateAddr string, uid int64, shortVer uint32, remoteAddr net.Addr) (*session.Session, error) {
	n.mu.RLock()
	s, found := n.sessions[sid]
	n.mu.RUnlock()
	if !found {
		conns, err := n.rpcClient.getConnPool(gateAddr)
		if err != nil {
			return nil, err
		}
		serializers := message.ReadSerializers()
		ac := &acceptor{
			sid:         sid,
			gateClient:  clusterpb.NewMemberClient(conns.Get()),
			rpcHandler:  n.handler.processMessage,
			gateAddr:    gateAddr,
			serializers: serializers,
			remoteAddr:  remoteAddr,
		}
		s = session.New(ac, sid)

		n.handler.mu.RLock()
		version := n.handler.versionDict[shortVer]
		n.handler.mu.RUnlock()
		s.BindShortVer(shortVer)
		s.BindVersion(version)

		s.BindUID(uid)
		ac.session = s
		n.mu.Lock()
		n.sessions[sid] = s
		n.mu.Unlock()

		session.Inited(s)
	}
	return s, nil
}

// HandleRequest is called by grpc `HandleRequest`
func (n *Node) HandleRequest(_ context.Context, req *clusterpb.RequestMessage) (*clusterpb.MemberHandleResponse, error) {
	handler, found := n.handler.localHandlers[req.Route]
	if !found {
		return nil, fmt.Errorf("service not found in current node: %v", req.Route)
	}
	remoteAddr := &NetAddr{network: req.RemoteAddr.Network, addr: req.RemoteAddr.Addr}
	s, err := n.findOrCreateSession(req.SessionID, req.GateAddr, req.UID, req.ShortVer, remoteAddr)
	if err != nil {
		return nil, err
	}
	msg := &message.Message{
		Type:     message.Request,
		ShortVer: s.ShortVer(),
		ID:       req.ID,
		Route:    req.Route,
		Data:     req.Data,
	}
	n.handler.localProcess(handler, req.ID, s, msg)
	return &clusterpb.MemberHandleResponse{}, nil
}

// HandleNotify is called by grpc `HandleNotify`
func (n *Node) HandleNotify(_ context.Context, req *clusterpb.NotifyMessage) (*clusterpb.MemberHandleResponse, error) {
	handler, found := n.handler.localHandlers[req.Route]
	if !found {
		return nil, fmt.Errorf("service not found in current node: %v", req.Route)
	}
	remoteAddr := &NetAddr{network: req.RemoteAddr.Network, addr: req.RemoteAddr.Addr}
	s, err := n.findOrCreateSession(req.SessionID, req.GateAddr, req.UID, req.ShortVer, remoteAddr)
	if err != nil {
		return nil, err
	}
	msg := &message.Message{
		Type:     message.Notify,
		ShortVer: s.ShortVer(),
		ID:       req.ID,
		Route:    req.Route,
		Data:     req.Data,
	}
	n.handler.localProcess(handler, req.ID, s, msg)
	return &clusterpb.MemberHandleResponse{}, nil
}

// HandlePush is called by grpc `HandlePush`
func (n *Node) HandlePush(_ context.Context, req *clusterpb.PushMessage) (*clusterpb.MemberHandleResponse, error) {
	s := n.findSession(req.SessionID)
	if s == nil {
		return &clusterpb.MemberHandleResponse{}, fmt.Errorf("session not found: %v", req.SessionID)
	}
	return &clusterpb.MemberHandleResponse{}, s.Push(req.Route, req.Data)
}

// HandleResponse is called by grpc `HandleResponse`
func (n *Node) HandleResponse(_ context.Context, req *clusterpb.ResponseMessage) (*clusterpb.MemberHandleResponse, error) {
	s := n.findSession(req.SessionID)
	if s == nil {
		return &clusterpb.MemberHandleResponse{}, fmt.Errorf("session not found: %v", req.SessionID)
	}
	return &clusterpb.MemberHandleResponse{}, s.ResponseMid(req.ID, req.Route, req.Data)
}

// NewMember is called by grpc `NewMember`
func (n *Node) NewMember(_ context.Context, req *clusterpb.NewMemberRequest) (*clusterpb.NewMemberResponse, error) {
	n.handler.addMember(req.MemberInfo)
	n.cluster.addMember(req.MemberInfo)
	return &clusterpb.NewMemberResponse{}, nil
}

// DelMember is called by grpc `DelMember`
func (n *Node) DelMember(_ context.Context, req *clusterpb.DelMemberRequest) (*clusterpb.DelMemberResponse, error) {
	n.handler.delMember(req.ServiceAddr)
	n.cluster.delMember(req.ServiceAddr)
	return &clusterpb.DelMemberResponse{}, nil
}

// SessionClosed implements the MemberServer interface
func (n *Node) SessionClosed(_ context.Context, req *clusterpb.SessionClosedRequest) (*clusterpb.SessionClosedResponse, error) {
	n.mu.Lock()
	s, found := n.sessions[req.SessionID]
	delete(n.sessions, req.SessionID)
	n.mu.Unlock()
	if found {
		session.Closed(s)
	}
	return &clusterpb.SessionClosedResponse{}, nil
}

// CloseSession implements the MemberServer interface
func (n *Node) CloseSession(_ context.Context, req *clusterpb.CloseSessionRequest) (*clusterpb.CloseSessionResponse, error) {
	n.mu.Lock()
	s, found := n.sessions[req.SessionID]
	delete(n.sessions, req.SessionID)
	n.mu.Unlock()
	if found {
		s.Close()
	}
	return &clusterpb.CloseSessionResponse{}, nil
}

// PerformConvention implements the MemberServer interface
func (n *Node) PerformConvention(_ context.Context, req *clusterpb.PerformConventionRequest) (*clusterpb.PerformConventionResponse, error) {
	if n.conventioner.acceptor != nil {
		data, err := n.conventioner.acceptor.React(req.Sig, req.Data)
		if err != nil {
			return &clusterpb.PerformConventionResponse{}, fmt.Errorf("member %s react error %s", n.Label, err.Error())
		}
		return &clusterpb.PerformConventionResponse{Label: n.Label, Data: data}, nil
	}
	return &clusterpb.PerformConventionResponse{}, nil
}
