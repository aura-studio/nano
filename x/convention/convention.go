package convention

import (
	"encoding/json"

	"github.com/lonng/nano/x/stats"
	"github.com/lonng/nano/cluster"
)

const (
	// RuntimeStats collects other members' runtime stats
	RuntimeStats int64 = iota
	// MessageStats collects gate server's message stats
	MessageStats
)

// MemberConvention implements Convention
// applying a easy version for convention
type MemberConvention struct {
	cluster.Convention
	Transmitter cluster.Transmitter
	Acceptor    *MemberConvention
}

// NewMemberConvention creates a new member convention
func NewMemberConvention() *MemberConvention {
	m := &MemberConvention{}
	m.Acceptor = m
	return m
}

// Establish implements Convention.Establish
func (m *MemberConvention) Establish(t cluster.Transmitter) cluster.Acceptor {
	m.Transmitter = t
	return m
}

// MemberConvention gets current node
func (m *MemberConvention) Node() *cluster.Node {
	return m.Transmitter.Node()
}

// Unicast implements Transmitter.Unicast
func (m *MemberConvention) Unicast(label string, sig int64, msg []byte) ([]byte, error) {
	return m.Transmitter.Unicast(label, sig, msg)
}

// Multicast implements Transmitter.Multicast
func (m *MemberConvention) Multicast(sig int64, msg []byte) ([]string, [][]byte, error) {
	return m.Transmitter.Multicast(sig, msg)
}

// React implements Acceptor.React
func (m *MemberConvention) React(sig int64, msg []byte) ([]byte, error) {
	switch sig {
	case RuntimeStats:
		return m.RuntimeStats(msg)
	case MessageStats:
		return m.MessageStats(msg)
	}
	return nil, nil
}

// RuntimeStats generates runtime stats info
func (m *MemberConvention) RuntimeStats(msg []byte) ([]byte, error) {
	return json.Marshal(stats.NewRuntimeStats().Info())
}

func (m *MemberConvention) MessageStats(msg []byte) ([]byte, error) {
	return json.Marshal(stats.NewMessageStats().Info())
}
