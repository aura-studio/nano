package plain

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/aura-studio/nano/codec"
	"github.com/aura-studio/nano/env"
	"github.com/aura-studio/nano/packet"
)

// Codec constants.
const (
	HeadLength    = 4
	MaxPacketSize = 64 * 1024
)

type Codec struct {
}

func NewCodec() *Codec {
	return &Codec{}
}

func (c *Codec) Encoder() codec.Encoder {
	return NewEncoder()
}

func (c *Codec) Decoder() codec.Decoder {
	return NewDecoder()
}

type Encoder struct {
	buf *bytes.Buffer
}

func NewEncoder() *Encoder {
	return &Encoder{
		buf: bytes.NewBuffer(nil),
	}
}

func (e *Encoder) Encode(packets ...*packet.Packet) ([]byte, error) {
	for _, p := range packets {
		err := binary.Write(e.buf, binary.BigEndian, uint32(p.Length))
		if err != nil {
			e.buf.Reset()
			return nil, err
		}
		e.buf.Write(p.Data)
	}
	data := e.buf.Next(e.buf.Len())
	e.buf.Reset()
	return data, nil
}

// ErrPacketSizeExcced is the error used for encode/decode.
var ErrPacketSizeExcced = errors.New("codec: packet size exceed")

// A Decoder reads and decodes network data slice
type Decoder struct {
	buf  *bytes.Buffer
	size int // last packet length
}

// NewDecoder returns a new decoder that used for decode network bytes slice.
func NewDecoder() *Decoder {
	return &Decoder{
		buf:  bytes.NewBuffer(nil),
		size: -1,
	}
}

func (d *Decoder) forward() error {
	header := d.buf.Next(HeadLength)
	d.size = int(binary.BigEndian.Uint32(header[:]))

	// packet length limitation
	if env.Safe && d.size > MaxPacketSize {
		return ErrPacketSizeExcced
	}

	return nil
}

// Decode decode the network bytes slice to packet.Packet(s)
func (d *Decoder) Decode(data []byte) ([]*packet.Packet, error) {
	d.buf.Write(data)

	var (
		packets []*packet.Packet
		err     error
	)
	// check length
	if d.buf.Len() < HeadLength {
		return nil, err
	}

	// first time
	if d.size < 0 {
		if err = d.forward(); err != nil {
			return nil, err
		}
	}

	for d.size <= d.buf.Len() {
		p := &packet.Packet{Length: d.size, Data: d.buf.Next(d.size)}
		packets = append(packets, p)

		// more packet
		if d.buf.Len() < HeadLength {
			d.size = -1
			break
		}

		if err = d.forward(); err != nil {
			return nil, err
		}

	}

	return packets, nil
}
