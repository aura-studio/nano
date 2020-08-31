package cluster

import (
	"context"
	"fmt"

	"github.com/lonng/nano/cluster/clusterpb"
)

type (
	// Transmitter unicasts & multicasts msg to
	Transmitter interface {
		Unicast(label string, sig int64, msg []byte) ([]byte, error)
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
	if node.Convention == nil {
		return nil
	}
	transmitter := &transmitter{
		node: node,
	}
	acceptor := node.Convention.Establish(transmitter)
	return &conventioner{
		transmitter: transmitter,
		acceptor:    acceptor,
	}
}

// Label returns current node label
func (t *transmitter) Label() string {
	return t.node.Label
}

// Unicast implements Transmitter.Unicast
func (t *transmitter) Unicast(label string, sig int64, msg []byte) ([]byte, error) {
	request := &clusterpb.PerformConventionRequest{Sig: sig, Data: msg}

	members := t.node.cluster.remoteMemebers()
	remote, ok := members[label]
	if !ok {
		return nil, fmt.Errorf("member not found by label %s", label)
	}
	pool, err := t.node.rpcClient.getConnPool(remote)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve connection pool for address %s %v", remote, err)
	}
	client := clusterpb.NewMemberClient(pool.Get())
	response, err := client.PerformConvention(context.Background(), request)
	if err != nil {
		return nil, fmt.Errorf("cannot perform convention in remote address %s %v", remote, err)
	}

	return response.Data, nil
}

// Unicast implements Transmitter.Multicast
func (t *transmitter) Multicast(sig int64, msg []byte) ([]string, [][]byte, error) {
	var labels []string
	var data [][]byte
	request := &clusterpb.PerformConventionRequest{Sig: sig, Data: msg}
	members := t.node.cluster.remoteMemebers()
	for _, remote := range members {
		pool, err := t.node.rpcClient.getConnPool(remote)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot retrieve connection pool for address %s %v", remote, err)
		}
		client := clusterpb.NewMemberClient(pool.Get())
		resp, err := client.PerformConvention(context.Background(), request)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot perform convention in remote address %s %v", remote, err)
		}
		labels = append(labels, resp.Label)
		data = append(data, resp.Data)
	}
	return labels, data, nil
}
