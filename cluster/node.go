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
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aura-studio/nano/cluster/clusterpb"
	"github.com/aura-studio/nano/env"
	"github.com/aura-studio/nano/log"
	"github.com/aura-studio/nano/message"
	"github.com/aura-studio/nano/options"
	"github.com/aura-studio/nano/session"
	"github.com/aura-studio/nano/style"
	"github.com/aura-studio/nano/upgrader/httpupgrader"
	"github.com/aura-studio/nano/upgrader/wsupgrader"
	"github.com/aura-studio/snowflake"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	grpcresolver "google.golang.org/grpc/resolver"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	etcdresolver "go.etcd.io/etcd/client/v3/naming/resolver"

	_ "net/http/pprof"
)

// Node represents a node in nano cluster, which will contains a group of services.
// All services will register to cluster and messages will be forwarded to the node
// which provides respective service
type Node struct {
	*options.Options        // current node options
	NodeID           uint32 // current node ID

	cluster   *cluster
	handler   *LocalHandler
	server    *grpc.Server
	rpcClient *rpcClient

	etcdClient   *clientv3.Client
	etcdManagers map[string]endpoints.Manager
	grpcResolver grpcresolver.Builder

	mu       sync.RWMutex
	sessions map[uint64]*session.Session
	state    uint32
}

// Startup bootstraps a start up.
func (n *Node) Startup() error {
	n.setServerID()

	n.etcdManagers = map[string]endpoints.Manager{}
	n.sessions = map[uint64]*session.Session{}
	n.cluster = newCluster(n)
	n.handler = newHandler(n)
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

	if n.TCPAddr != "" {
		go n.listenAndServeTCP()
	}

	if n.HttpAddr != "" {
		go n.listenAndServeHttp()
	}

	return nil
}

func (n *Node) setServerID() {
	if n.Etcd {
		epoch, err := time.Parse("2006-01-02 15:04:05", "2022-02-22 22:22:22")
		if err != nil {
			log.Fatal(err)
		}
		pattern := snowflake.NewPattern(
			epoch, time.Millisecond,
			[2]uint8{0, 10},  // 1,024,000 iops
			[2]uint8{51, 12}, // 4096 nodes
			[2]uint8{10, 41}, // 68 Years
		)
		etcd := &snowflake.Etcd{
			Prefix:        "/ServerID/",
			Addr:          n.AdvertiseAddr,
			Timeout:       10 * time.Second,
			LeaseTime:     time.Minute,
			LeaseInterval: time.Minute - 10*time.Second,
		}
		env.SnowflakeNode, err = pattern.NewNode(snowflake.WithEtcdNode(etcd))
		if err != nil {
			log.Fatal(err)
		}
		n.NodeID = uint32(env.SnowflakeNode.Node())
	} else {
		if n.MemberAddr == "" {
			n.NodeID = 0
			return
		}

		parts := strings.Split(n.MemberAddr, ":")
		var host string
		if parts[0] == "" {
			host = "0.0.0.0"
		} else {
			host = parts[0]
		}
		port, _ := strconv.Atoi(parts[1])
		addrs, _ := net.LookupHost(host)
		var nodeID = uint32(0)
		for _, addr := range addrs {
			bits := strings.Split(addr, ".")
			if len(bits) != 4 {
				continue
			}
			b2, _ := strconv.Atoi(bits[2])
			b3, _ := strconv.Atoi(bits[3])
			var sum uint32
			sum += uint32(b2) << 24
			sum += uint32(b3) << 16
			sum += uint32(port)
			if sum > nodeID {
				nodeID = sum
			}
		}

		if nodeID == 0 {
			nodeID = rand.Uint32()
		}

		n.NodeID = nodeID
	}
}

func (n *Node) WholeInterface(addr string) string {
	return addr[strings.Index(addr, ":"):]
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

	listener, err := net.Listen("tcp", n.WholeInterface(n.MemberAddr))
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

	// etcd mode
	if n.Etcd {
		n.etcdClient, err = clientv3.New(clientv3.Config{
			Endpoints:   strings.Split(n.AdvertiseAddr, ","), // etcd address
			DialTimeout: 10 * time.Second,
		})

		for _, service := range n.handler.LocalService() {
			etcdManager, err := endpoints.NewManager(n.etcdClient, service)
			if err != nil {
				return err
			}
			n.etcdManagers[service] = etcdManager
			if err := etcdManager.AddEndpoint(n.etcdClient.Ctx(),
				fmt.Sprintf("%s/%s/%s", service, env.Version, n.MemberAddr), endpoints.Endpoint{Addr: n.MemberAddr}); err != nil {
				return err
			}
		}
		n.grpcResolver, err = etcdresolver.NewBuilder(n.etcdClient)
		if err != nil {
			return err
		}
		env.GrpcOptions = append(env.GrpcOptions, grpc.WithResolvers(n.grpcResolver)) // load balancer

	} else {
		if n.IsMaster {
			clusterpb.RegisterMasterServer(n.server, n.cluster)
			member := &Member{
				isMaster: true,
				memberInfo: &clusterpb.MemberInfo{
					Label:       n.Label,
					Version:     env.Version,
					ServiceAddr: n.MemberAddr,
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
					ServiceAddr: n.MemberAddr,
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
	}

	return nil
}

// Close all components registered by application, that
// call by reverse order against register
func (n *Node) Close() {
	atomic.AddUint32(&n.state, 1)

	for {
		select {
		case env.ConnDie <- true:
		case <-time.After(time.Second):
			time.After(time.Second)
			goto CLOSE
		}
	}
CLOSE:
	// reverse call `BeforeClose` hooks
	components := n.Components.List()
	length := len(components)
	for i := length - 1; i >= 0; i-- {
		components[i].Comp.BeforeClose()
	}

	// reverse call `Close` hooks
	for i := length - 1; i >= 0; i-- {
		components[i].Comp.Close()
	}

	// etcd mode
	if n.Etcd {
		for service, etcdManager := range n.etcdManagers {
			etcdManager.DeleteEndpoint(n.etcdClient.Ctx(), service+"/"+n.MemberAddr)
		}
		n.etcdClient.Close()
	} else {
		if !n.IsMaster && n.AdvertiseAddr != "" {
			pool, err := n.rpcClient.getConnPool(n.AdvertiseAddr)
			if err != nil {
				log.Errorln("Retrieve master address error", err)
				goto EXIT
			}
			client := clusterpb.NewMasterClient(pool.Get())
			request := &clusterpb.UnregisterRequest{
				ServiceAddr: n.MemberAddr,
			}
			_, err = client.Unregister(context.Background(), request)
			if err != nil {
				log.Errorln("Unregister current node failed", err)
				goto EXIT
			}
		}
	}

EXIT:
	if n.server != nil {
		n.server.GracefulStop()
	}
}

// Enable current server accept connection
func (n *Node) listenAndServeTCP() {
	listener, err := net.Listen("tcp", n.WholeInterface(n.TCPAddr))
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

		if atomic.LoadUint32(&n.state) != 0 {
			continue
		}

		go n.handler.handle(conn)
	}
}

func (n *Node) listenAndServeHttp() {
	router := mux.NewRouter()
	router.HandleFunc("/{service:[A-Za-z0-9\\-]*}/{handler:[A-Za-z0-9\\-]*}", func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		service := strings.Join(style.GoogleChain(params["service"]), "")
		handler := strings.Join(style.GoogleChain(params["handler"]), "")
		route := fmt.Sprintf("%s.%s", service, handler)
		params["route"] = route
		n.routerHandler(params, w, r)
	})
	router.HandleFunc("/{route:[A-Za-z0-9\\.]*}", func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		n.routerHandler(params, w, r)
	})
	http.Handle("/", router)

	addr := n.WholeInterface(n.HttpAddr)

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

func (n *Node) routerHandler(params map[string]string, w http.ResponseWriter, r *http.Request) {
	var (
		conn net.Conn
		err  error
	)

	if route, ok := params["route"]; ok && route == "websocket" {
		conn, err = wsupgrader.NewUpgrader().Upgrade(w, r, params)
	} else {
		conn, err = httpupgrader.NewUpgrader().Upgrade(w, r, params)
	}

	if err != nil {
		log.Errorf("Upgrade failure, connection will be dropped. URI=%s, Error=%s", r.RequestURI, err.Error())
		if conn != nil {
			conn.Close()
		}
		return
	}

	if atomic.LoadUint32(&n.state) != 0 {
		log.Info("Server is closing, connection will be dropped.")
		if conn != nil {
			conn.Close()
		}
		return
	}

	go n.handler.handle(conn)
}

func (n *Node) ListenAndServeDebug() {
	if err := http.ListenAndServe(n.WholeInterface(n.DebugAddr), nil); err != nil {
		log.Fatal(err.Error())
	}
}

func (n *Node) CountSessions() int {
	n.mu.Lock()
	c := len(n.sessions)
	n.mu.Unlock()
	return c
}

func (n *Node) storeSession(s *session.Session) {
	n.mu.Lock()
	n.sessions[s.ID()] = s
	n.mu.Unlock()
}

func (n *Node) findSession(sid uint64) *session.Session {
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

func (n *Node) findOrCreateSession(sid uint64, gateAddr string, uid int64, shortVer uint32, remoteAddr net.Addr, branch uint32, dataType uint32) (*session.Session, error) {
	n.mu.RLock()
	s, found := n.sessions[sid]
	n.mu.RUnlock()
	if !found {
		conns, err := n.rpcClient.getConnPool(gateAddr)
		if err != nil {
			return nil, err
		}
		ac := &acceptor{
			sid:        sid,
			gateClient: clusterpb.NewMemberClient(conns.Get()),
			rpcHandler: n.handler.processMessage,
			gateAddr:   gateAddr,
			remoteAddr: remoteAddr,
		}
		s = session.New(ac, sid)

		n.handler.mu.RLock()
		version := n.handler.versionDict[shortVer]
		n.handler.mu.RUnlock()
		s.BindShortVer(shortVer)
		s.BindVersion(version)
		s.BindBranch(branch)
		s.DataType.Store(dataType)

		s.VersionBound = true

		s.BindUID(uid)
		ac.session = s
		n.mu.Lock()
		n.sessions[sid] = s
		n.mu.Unlock()

		session.Inited(s)

		go func() {
			defer func() {
				if err := recover(); err != nil {
					log.Errorln("Session goroutine panic", err)
				}
			}()

			_, err = ac.gateClient.SessionCreated(context.Background(), &clusterpb.SessionCreatedRequest{
				Addr:      n.MemberAddr,
				SessionID: sid,
			})
		}()
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
	s, err := n.findOrCreateSession(req.SessionID, req.GateAddr, req.UID, req.ShortVer, remoteAddr, req.Branch, req.DataType)
	if err != nil {
		return nil, err
	}
	msg := &message.Message{
		Type:     message.Request,
		Branch:   s.Branch(),
		ShortVer: s.ShortVer(),
		ID:       req.ID,
		Route:    req.Route,
		DataType: req.DataType,
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
	s, err := n.findOrCreateSession(req.SessionID, req.GateAddr, req.UID, req.ShortVer, remoteAddr, req.Branch, req.DataType)
	if err != nil {
		return nil, err
	}
	msg := &message.Message{
		Type:     message.Notify,
		Branch:   s.Branch(),
		ShortVer: s.ShortVer(),
		ID:       req.ID,
		Route:    req.Route,
		DataType: req.DataType,
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
	return &clusterpb.MemberHandleResponse{}, session.Context(s, req.ID).Response(req.Route, req.Data)
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

// SessionCreated implements the MemberServer interface
func (n *Node) SessionCreated(_ context.Context, req *clusterpb.SessionCreatedRequest) (*clusterpb.SessionCreatedResponse, error) {
	s := n.findSession(req.SessionID)
	if s != nil {
		s.AddRemoteSessionAddr(req.Addr)
	} else {
		go func() {
			defer func() {
				if err := recover(); err != nil {
					log.Errorf("session created panic: %v", err)
				}
			}()

			pool, err := n.rpcClient.getConnPool(req.Addr)
			if err != nil {
				log.Error(err)
			}
			client := clusterpb.NewMemberClient(pool.Get())
			_, err = client.SessionClosed(context.Background(), &clusterpb.SessionClosedRequest{
				SessionID: req.SessionID,
			})
			if err != nil {
				log.Error(err)
			}
		}()
	}
	return &clusterpb.SessionCreatedResponse{}, nil
}
