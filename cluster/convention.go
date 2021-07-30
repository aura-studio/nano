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
	var labels []string
	var dataList [][]byte

	for _, addr := range t.addrs(label) {
		label, data, err := t.invoke(addr, sig, msg)
		if err != nil {
			return nil, err
		}
		labels = append(labels, label)
		dataList = append(dataList, data)
	}

	return dataList, nil
}

// Unicast implements Transmitter.Multicast
func (t *transmitter) Multicast(sig int64, msg []byte) ([]string, [][]byte, error) {
	var labels []string
	var dataList [][]byte

	for _, addr := range t.addrs("") {
		label, data, err := t.invoke(addr, sig, msg)
		if err != nil {
			return nil, nil, err
		}
		labels = append(labels, label)
		dataList = append(dataList, data)
	}

	return labels, dataList, nil
}

func (t *transmitter) addrs(label string) []string {
	var addrs []string
	if label == "" || label == t.node.Label {
		addrs = append(addrs, t.node.ServiceAddr)
	}
	for _, member := range t.node.cluster.members {
		if label == "" || member.memberInfo.Label == label {
			addrs = append(addrs, member.memberInfo.ServiceAddr)
		}
	}
	return addrs
}

func (t *transmitter) invoke(addr string, sig int64, data []byte) (string, []byte, error) {
	request := &clusterpb.PerformConventionRequest{Sig: sig, Data: data}
	pool, err := t.node.rpcClient.getConnPool(addr)
	if err != nil {
		return "", nil, fmt.Errorf("cannot retrieve connection pool for address %s %v", addr, err)
	}
	client := clusterpb.NewMemberClient(pool.Get())
	response, err := client.PerformConvention(context.Background(), request)
	if err != nil {
		return "", nil, fmt.Errorf("cannot perform convention in remote address %s %v", addr, err)
	}
	return response.Label, request.Data, nil
}
