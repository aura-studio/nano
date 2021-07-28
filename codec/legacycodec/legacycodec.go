package legacycodec

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
	msgHeadLength      = 16
	msgContentTypeMask = 0x1
)

var (
	ErrPacketSizeExcced = errors.New("codec: packet size exceed")
	ErrWrongMessageType = errors.New("codec: wrong message type")
	ErrInvalidMessage   = errors.New("codec: invalid message")
)

type CodecEntity struct {
	dictionary message.Dictionary
	writeBuf   *bytes.Buffer
	readBuf    *bytes.Buffer
	size       int // last packet length
}

func NewCodecEntity(dictionary message.Dictionary) *CodecEntity {
	return &CodecEntity{
		dictionary: dictionary,
		writeBuf:   bytes.NewBuffer(nil),
		readBuf:    bytes.NewBuffer(nil),
		size:       -1,
	}
}

func (c *CodecEntity) EncodePacket(packets []*packet.Packet) ([]byte, error) {
	var length int
	for _, p := range packets {
		err := binary.Write(c.writeBuf, binary.LittleEndian, uint16(p.Length+2))
		if err != nil {
			return nil, err
		}
		length += 2
		c.writeBuf.Write(p.Data)
		length += len(p.Data)
	}
	data := c.writeBuf.Next(length)

	return data, nil
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
		p := &packet.Packet{Length: c.size, Data: c.readBuf.Next(c.size)}
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
	if !m.TypeValid() {
		return nil, ErrWrongMessageType
	}
	var offset uint64 = 0
	buf := make([]byte, 16)

	// encode version ID
	binary.LittleEndian.PutUint32(buf[offset:], m.ShortVer)
	offset += 4

	// encode msg ID
	binary.LittleEndian.PutUint32(buf[offset:], uint32(m.ID))
	offset += 4

	// encode compressed route ID
	code, err := c.dictionary.IndexRoute(m.Route)
	if err != nil {
		return nil, err
	}
	binary.LittleEndian.PutUint16(buf[offset:], uint16(code/65536))
	offset += 2
	binary.LittleEndian.PutUint16(buf[offset:], uint16(code%65536))
	offset += 2

	// encode data length
	length := uint16(len(m.Data))
	binary.LittleEndian.PutUint16(buf[offset:], length)
	offset += 2

	// encode flag
	flag := uint16(0)
	flag |= msgContentTypeMask
	binary.LittleEndian.PutUint16(buf[offset:], flag)

	// encode data
	buf = append(buf, m.Data...)

	return buf, nil
}

func (c *CodecEntity) DecodeMessage(data []byte) (*message.Message, error) {
	var err error
	if len(data) < msgHeadLength {
		return nil, ErrInvalidMessage
	}
	var offset uint64 = 0
	m := message.New()
	m.Type = message.Response

	// decode version ID
	m.ShortVer = binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	// decode msg ID
	m.ID = uint64(binary.LittleEndian.Uint32(data[offset:]))
	offset += 4

	// decode compressed route ID
	code := uint32(binary.LittleEndian.Uint16(data[offset:]))*65536 +
		uint32(binary.LittleEndian.Uint16(data[offset+2:]))
	m.Route, err = c.dictionary.IndexCode(code)
	if err != nil {
		return nil, err
	}

	m.Data = data[msgHeadLength:]

	return m, nil
}

type Codec struct {
	message.Dictionary
}

func NewCodec() *Codec {
	return &Codec{}
}

func (c *Codec) Entity(dictionary message.Dictionary) codec.CodecEntity {
	if dictionary == nil {
		dictionary = message.EmptyDictionary
	}
	return NewCodecEntity(dictionary)
}
