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
		Multicast(sig int64, msg []byte) ([][]byte, error)
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
	transmitter struct{}

	// Conventioner contains a transmitter & a acceptor
	Conventioner struct {
		Transmitter Transmitter
		Acceptor    Acceptor
	}
)

// NewConventioner creates a new conventioner
func NewConventioner(convention Convention) Conventioner {
	var transmitter = &transmitter{}
	acceptor := convention.Establish(transmitter)

	return Conventioner{
		Transmitter: transmitter,
		Acceptor:    acceptor,
	}
}

// Unicast implements Transmitter.Unicast
func (t *transmitter) Unicast(label string, sig int64, msg []byte) ([]byte, error) {
	request := &clusterpb.PerformConventionRequest{Sig: sig, Data: msg}

	members := CurrentNode.cluster.remoteMemebers()
	remote, ok := members[label]
	if !ok {
		return nil, fmt.Errorf("member not found by label %s", label)
	}
	pool, err := CurrentNode.rpcClient.getConnPool(remote)
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
func (t *transmitter) Multicast(sig int64, msg []byte) ([][]byte, error) {
	var data [][]byte
	request := &clusterpb.PerformConventionRequest{Sig: sig, Data: msg}
	members := CurrentNode.cluster.remoteMemebers()
	for _, remote := range members {
		pool, err := CurrentNode.rpcClient.getConnPool(remote)
		if err != nil {
			return nil, fmt.Errorf("cannot retrieve connection pool for address %s %v", remote, err)
		}
		client := clusterpb.NewMemberClient(pool.Get())
		resp, err := client.PerformConvention(context.Background(), request)
		if err != nil {
			return nil, fmt.Errorf("cannot perform convention in remote address %s %v", remote, err)
		}
		data = append(data, resp.Data)
	}
	return data, nil
}
