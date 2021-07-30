package cluster

import (
	"context"
	"fmt"

	"github.com/aura-studio/nano/cluster/clusterpb"
)

type (
	// Transmitter unicasts & multicasts msg to
	Transmitter interface {
		Node() *Node
		Unicast(label string, sig int64, msg []byte) ([][]byte, error)
		Multicast(sig int64, msg []byte) ([]string, [][]byte, error)
	}

	// Acceptor
	Acceptor interface {
		React(sig int64, msg []byte) ([]byte, error)
	}

	// Convention establish a connection
	Convention interface {
		Establish(Transmitter) Acceptor
	}

	// transmitter is to implement Transmitter
	transmitter struct {
		node *Node
	}

	// conventioner contains a transmitter & a acceptor
	conventioner struct {
		transmitter Transmitter
		acceptor    Acceptor
	}
)

// newConventioner creates a new conventioner
func newConventioner(node *Node) *conventioner {
	transmitter := &transmitter{
		node: node,
	}
	var acceptor Acceptor
	if node.Convention != nil {
		acceptor = node.Convention.Establish(transmitter)
	}
	return &conventioner{
		transmitter: transmitter,
		acceptor:    acceptor,
	}
}

// Node returns current node
func (t *transmitter) Node() *Node {
	return t.node
}

// Unicast implements Transmitter.Unicast
func (t *transmitter) Unicast(label string, sig int64, msg []byte) ([][]byte, error) {
	request := &clusterpb.PerformConventionRequest{Sig: sig, Data: msg}
	var data [][]byte
	for _, member := range t.node.cluster.members {
		if member.memberInfo.Label == label {
			addr := member.memberInfo.ServiceAddr
			pool, err := t.node.rpcClient.getConnPool(addr)
			if err != nil {
				return nil, fmt.Errorf("cannot retrieve connection pool for address %s %v", addr, err)
			}
			client := clusterpb.NewMemberClient(pool.Get())
			resp, err := client.PerformConvention(context.Background(), request)
			if err != nil {
				return nil, fmt.Errorf("cannot perform convention in remote address %s %v", addr, err)
			}
			data = append(data, resp.Data)
		}
	}

	return data, nil
}

// Unicast implements Transmitter.Multicast
func (t *transmitter) Multicast(sig int64, msg []byte) ([]string, [][]byte, error) {
	var labels []string
	var data [][]byte
	request := &clusterpb.PerformConventionRequest{Sig: sig, Data: msg}
	for _, member := range t.node.cluster.members {
		addr := member.memberInfo.ServiceAddr
		pool, err := t.node.rpcClient.getConnPool(addr)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot retrieve connection pool for address %s %v", addr, err)
		}
		client := clusterpb.NewMemberClient(pool.Get())
		resp, err := client.PerformConvention(context.Background(), request)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot perform convention in remote address %s %v", addr, err)
		}
		labels = append(labels, resp.Label)
		data = append(data, resp.Data)
	}
	return labels, data, nil
}
