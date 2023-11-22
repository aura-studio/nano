package proxycodec

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/aura-studio/nano/codec"
	"github.com/aura-studio/nano/env"
	"github.com/aura-studio/nano/message"
	"github.com/aura-studio/nano/packet"
)

const (
	HeadLength    = 2
	MaxPacketSize = 64 * 1024
)

const (
	msgHeadLength = 16
)

const (
	route = "Proxy.TransferBytes"
)

var (
	ErrPacketSizeExcced = errors.New("codec: packet size exceed")
	ErrWrongMessageType = errors.New("codec: wrong message type")
	ErrInvalidMessage   = errors.New("codec: invalid message")
)

type CodecEntity struct {
	writeBuf   *bytes.Buffer
	readBuf    *bytes.Buffer
	size       int // last packet length
	dictionary message.Dictionary
}

func NewCodecEntity() *CodecEntity {
	return &CodecEntity{
		writeBuf: bytes.NewBuffer(nil),
		readBuf:  bytes.NewBuffer(nil),
		size:     -1,
	}
}

func (c *CodecEntity) EncodePacket(packets []*packet.Packet) ([]byte, error) {
	var length int
	for _, p := range packets {
		c.writeBuf.Write(p.Data)
		length += len(p.Data)
	}
	data := c.writeBuf.Next(length)

	copyData := make([]byte, len(data))
	copy(copyData, data)

	return copyData, nil
}

func (c *CodecEntity) DecodePacket(data []byte) ([]*packet.Packet, error) {
	forward := func() error {
		header := c.readBuf.Next(HeadLength)
		c.size = int(binary.LittleEndian.Uint16(header)) - HeadLength

		// packet length limitation
		if env.Safe && (c.size > MaxPacketSize || c.size < 0) {
			return ErrPacketSizeExcced
		}

		return nil
	}

	var (
		packets []*packet.Packet
		err     error
	)

	c.readBuf.Write(data)

	// check length
	if c.readBuf.Len() < HeadLength {
		return nil, err
	}

	if c.size < 0 {
		if err = forward(); err != nil {
			return nil, err
		}
	}

	for c.size <= c.readBuf.Len() {
		data := make([]byte, c.size+HeadLength)
		binary.LittleEndian.PutUint16(data, uint16(c.size)+HeadLength)
		copy(data[HeadLength:], c.readBuf.Next(c.size))
		p := &packet.Packet{
			Length: c.size,
			Data:   data,
		}
		packets = append(packets, p)

		// more packet
		if c.readBuf.Len() < HeadLength {
			c.size = -1
			break
		}

		if err = forward(); err != nil {
			return nil, err
		}
	}

	return packets, nil
}

func (c *CodecEntity) EncodeMessage(m *message.Message) ([]byte, error) {
	return m.Data, nil
}

func (c *CodecEntity) DecodeMessage(data []byte) (*message.Message, error) {
	if len(data) < msgHeadLength {
		return nil, ErrInvalidMessage
	}
	var offset uint64 = 0
	m := message.New()
	offset += 2
	m.Type = message.Request

	// decode version ID
	m.ShortVer = binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	// decode msg ID
	m.ID = uint64(binary.LittleEndian.Uint32(data[offset:]))

	m.Route = route

	m.Data = data

	return m, nil
}

type Codec struct {
}

func NewCodec() *Codec {
	return &Codec{}
}

func (c *Codec) Entity(message.Dictionary) codec.CodecEntity {
	return NewCodecEntity()
}
