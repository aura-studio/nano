package cluster

import (
	"context"
	"net"

	"github.com/aura-studio/nano/cluster/clusterpb"
	"github.com/aura-studio/nano/env"
	"github.com/aura-studio/nano/log"
	"github.com/aura-studio/nano/message"
	"github.com/aura-studio/nano/serialize"
	"github.com/aura-studio/nano/session"
)

type acceptor struct {
	sid         int64
	gateClient  clusterpb.MemberClient
	session     *session.Session
	lastMid     uint64
	rpcHandler  rpcHandler
	gateAddr    string
	serializers map[string]serialize.Serializer // copy system serializers for agent
	remoteAddr  net.Addr
}

// Push implements the session.NetworkEntity interface
func (a *acceptor) Push(route string, v interface{}) error {
	data, err := message.Serialize(v)
	if err != nil {
		return err
	}

	if env.Debug {
		switch d := v.(type) {
		case []byte:
			log.Infof("Type=Push, Route=%s, SSID=%d, SID=%d, Version=%s, UID=%d, MID=%d, Data=%dbytes",
				route, a.session.SSID(), a.session.SID(), a.session.Version(), a.session.UID(), 0, len(d))
		default:
			log.Infof("Type=Push, Route=%s, SSID=%d, SID=%d, Version=%s, UID=%d, MID=%d, Data=%+v",
				route, a.session.SSID(), a.session.SID(), a.session.Version(), a.session.UID(), 0, v)
		}
	}

	request := &clusterpb.PushMessage{
		SessionID: a.sid,
		ShortVer:  a.session.ShortVer(),
		Route:     route,
		Data:      data,
	}
	_, err = a.gateClient.HandlePush(context.Background(), request)
	return err
}

// RPC implements the session.NetworkEntity interface
func (a *acceptor) RPC(route string, v interface{}) error {
	return a.RPCMid(a.lastMid, route, v)
}

// RPC implements the session.NetworkEntity interface
func (a *acceptor) RPCMid(mid uint64, route string, v interface{}) error {
	data, err := message.RouteSerialize(a.serializers, route, v)
	if err != nil {
		return err
	}

	if env.Debug {
		switch d := v.(type) {
		case []byte:
			log.Infof("Type=Notify, Route=%s, SSID=%d, SID=%d, Version=%s, UID=%d, MID=%d, Data=%dbytes",
				route, a.session.SSID(), a.session.SID(), a.session.Version(), a.session.UID(), mid, len(d))
		default:
			log.Infof("Type=Notify, Route=%s, SSID=%d, SID=%d, Version=%s, UID=%d, MID=%d, Data=%+v",
				route, a.session.SSID(), a.session.SID(), a.session.Version(), a.session.UID(), mid, v)
		}
	}

	msg := &message.Message{
		Type:     message.Notify,
		ShortVer: a.session.ShortVer(),
		ID:       mid,
		Route:    route,
		Data:     data,
	}

	a.rpcHandler(a.session, msg, true)
	return nil
}

// Response implements the session.NetworkEntity interface
func (a *acceptor) Response(route string, v interface{}) error {
	return a.ResponseMid(a.lastMid, route, v)
}

// ResponseMid implements the session.NetworkEntity interface
func (a *acceptor) ResponseMid(mid uint64, route string, v interface{}) error {
	data, err := message.Serialize(v)
	if err != nil {
		return err
	}

	if env.Debug {
		switch d := v.(type) {
		case []byte:
			log.Infof("Type=Response, Route=%s, SSID=%d, SID=%d, Version=%s, UID=%d, MID=%d, Data=%dbytes",
				route, a.session.SSID(), a.session.SID(), a.session.Version(), a.session.UID(), mid, len(d))
		default:
			log.Infof("Type=Response, Route=%s, SSID=%d, SID=%d, Version=%s, UID=%d, MID=%d, Data=%+v",
				route, a.session.SSID(), a.session.SID(), a.session.Version(), a.session.UID(), mid, v)
		}
	}

	request := &clusterpb.ResponseMessage{
		SessionID: a.sid,
		ShortVer:  a.session.ShortVer(),
		ID:        mid,
		Route:     route,
		Data:      data,
	}
	_, err = a.gateClient.HandleResponse(context.Background(), request)
	return err
}

// Close implements the session.NetworkEntity interface
func (a *acceptor) Close() error {
	request := &clusterpb.CloseSessionRequest{
		SessionID: a.sid,
	}
	_, err := a.gateClient.CloseSession(context.Background(), request)
	return err
}

// RemoteAddr implements the session.NetworkEntity interface
func (a *acceptor) RemoteAddr() net.Addr {
	return a.remoteAddr
}
