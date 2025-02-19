package cluster

import (
	"context"
	"net"

	"github.com/aura-studio/nano/cluster/clusterpb"
	"github.com/aura-studio/nano/env"
	"github.com/aura-studio/nano/log"
	"github.com/aura-studio/nano/message"
	"github.com/aura-studio/nano/session"
)

type acceptor struct {
	sid        uint64
	gateClient clusterpb.MemberClient
	session    *session.Session
	rpcHandler rpcHandler
	gateAddr   string
	remoteAddr net.Addr
}

// Push implements the session.NetworkEntity interface
func (a *acceptor) Push(route string, v interface{}) error {
	data, err := a.session.Serialize(v)
	if err != nil {
		return err
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

	request := &clusterpb.PushMessage{
		SessionID:  a.sid,
		ShortVer:   a.session.ShortVer(),
		Route:      route,
		DataType:   a.session.DataType.Load(),
		Data:       data,
		CryptoType: a.session.CryptoType.Load(),
	}
	_, err = a.gateClient.HandlePush(context.Background(), request)
	return err
}

// RPC implements the session.NetworkEntity interface
func (a *acceptor) RPC(mid uint64, route string, v interface{}) error {
	data, err := a.session.Serialize(v)
	if err != nil {
		return err
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

// Response implements the session.NetworkEntity interface
func (a *acceptor) Response(mid uint64, route string, v interface{}) error {
	data, err := a.session.Serialize(v)
	if err != nil {
		return err
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

	request := &clusterpb.ResponseMessage{
		SessionID:  a.sid,
		ShortVer:   a.session.ShortVer(),
		ID:         mid,
		Route:      route,
		DataType:   a.session.DataType.Load(),
		Data:       data,
		CryptoType: a.session.CryptoType.Load(),
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
