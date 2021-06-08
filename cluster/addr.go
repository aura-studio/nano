package cluster

type NetAddr struct {
	network string
	addr    string
}

func (n *NetAddr) Network() string {
	return n.network
}

func (n *NetAddr) String() string {
	return n.addr
}
