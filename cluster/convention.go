package cluster

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
		Establish(Transmitter, Acceptor)
	}

	// transmitter is to implement Transmitter
	transmitter struct {
		node *Node
	}

	// Conventioner contains a transmitter & a acceptor
	Conventioner struct {
		Transmitter Transmitter
		Acceptor    Acceptor
	}
)

func NewConventioner(convention Convention) Conventioner {
	var transmitter = &transmitter{
		node: CurrentNode,
	}
	var acceptor Acceptor
	convention.Establish(transmitter, acceptor)

	return Conventioner{
		Transmitter: transmitter,
		Acceptor:    acceptor,
	}
}

func (t *transmitter) Unicast(label string, sig int64, msg []byte) ([]byte, error) {
	return t.node.cluster.Unicast(label, sig, msg)
}

func (t *transmitter) Multicast(sig int64, msg []byte) ([][]byte, error) {
	return t.node.cluster.Multicast(sig, msg)
}
