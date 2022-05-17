package cluster

import (
	"context"
	"fmt"

	"github.com/aura-studio/nano/cluster/clusterpb"
	"github.com/aura-studio/nano/log"
)

type (
	// Transmitter unicasts & multicasts msg to
	Transmitter interface {
		Node() *Node
		Unicast(addr string, sig int64, msg []byte) ([]byte, error)
		Multicast(label string, sig int64, msg []byte) ([][]byte, error)
		Broadcast(sig int64, msg []byte) ([]string, [][]byte, error)
	}

	// Acceptor
	Acceptor interface {
		React(sig int64, msg []byte) ([]byte, error)
	}

	// Convention establish a connection
	Convention interface {
		Establish(Transmitter) Acceptor
	}

	transmitter struct {
		Acceptor
		node *Node
	}
)

// newTransmitter creates a new conventioner
func newTransmitter(node *Node) *transmitter {
	c := &transmitter{nil, node}
	if node.Convention == nil {
		return c
	}
	c.Acceptor = node.Convention.Establish(c)
	return c
}

// Node returns current node
func (t *transmitter) Node() *Node {
	return t.node
}

// Unicast implements func Transmitter.Unicast
func (t *transmitter) Unicast(addr string, sig int64, msg []byte) ([]byte, error) {
	if addr == t.node.MemberAddr {
		return t.React(sig, msg)
	}
	_, data, err := t.invoke(addr, sig, msg)
	return data, err
}

// Multicast implements func Transmitter.Multicast
func (t *transmitter) Multicast(label string, sig int64, msg []byte) ([][]byte, error) {
	var labels []string
	var dataList [][]byte

	if label == t.node.Label {
		data, err := t.React(sig, msg)
		if err != nil {
			return nil, err
		}
		dataList = append(dataList, data)
	}

	for _, addr := range t.addrs(label) {
		label, data, err := t.invoke(addr, sig, msg)
		if err != nil {
			log.Errorf("transmitter broadcast %v err: %v", addr, err)
			continue
		}
		labels = append(labels, label)
		dataList = append(dataList, data)
	}

	return dataList, nil
}

// Broadcast implements func Transmitter.Broadcast
func (t *transmitter) Broadcast(sig int64, msg []byte) ([]string, [][]byte, error) {
	var labels []string
	var dataList [][]byte

	data, err := t.React(sig, msg)
	if err != nil {
		return nil, nil, err
	}
	dataList = append(dataList, data)

	for _, addr := range t.addrs("") {
		label, data, err := t.invoke(addr, sig, msg)
		if err != nil {
			log.Errorf("transmitter broadcast %v err: %v", addr, err)
			continue
		}
		labels = append(labels, label)
		dataList = append(dataList, data)
	}

	return labels, dataList, nil
}

func (t *transmitter) addrs(label string) []string {
	var addrs []string
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
	return response.Label, response.Data, nil
}
