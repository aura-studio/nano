package cryptocodec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/aura-studio/nano/codec"
	"github.com/aura-studio/nano/crypto/aes"
	"github.com/aura-studio/nano/env"
	"github.com/aura-studio/nano/message"
	"github.com/aura-studio/nano/packet"
)

const (
	HeadLength    = 4
	MaxPacketSize = 64 * 1024
)

const (
	msgRouteNotCompressMask = 0x08 // 0000 1000
	msgTypeMask             = 0x07 // 0000 0111
	msgBranchMask           = 0xF0 // 1111 0000
	msgHeadLength           = 0x02
)

var (
	ErrPacketSizeExcced   = errors.New("codec: packet size exceed")
	ErrWrongMessageType   = errors.New("codec: wrong message type")
	ErrInvalidMessage     = errors.New("codec: invalid message")
	ErrInvalidRouteLength = errors.New("codec: invalid route length")
)

var (
	aesKey = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
	}
)

type CodecEntity struct {
	dictionary message.Dictionary
	writeBuf   *bytes.Buffer
	readBuf    *bytes.Buffer
	size       int          // last packet length
	compressed *atomic.Bool // whether to use compressed msg to client
}

func NewCodecEntity(dictionary message.Dictionary) *CodecEntity {
	return &CodecEntity{
		writeBuf:   bytes.NewBuffer(nil),
		readBuf:    bytes.NewBuffer(nil),
		size:       -1,
		dictionary: dictionary,
		compressed: &atomic.Bool{},
	}
}

func (c *CodecEntity) EncodePacket(packets []*packet.Packet) ([]byte, error) {
	var length int
	for _, p := range packets {
		err := binary.Write(c.writeBuf, binary.BigEndian, uint32(p.Length))
		if err != nil {
			return nil, err
		}
		length += 4
		c.writeBuf.Write(p.Data)
		length += len(p.Data)
	}
	data := c.writeBuf.Next(length)

	return data, nil
}

func (c *CodecEntity) DecodePacket(data []byte) ([]*packet.Packet, error) {
	forward := func() error {
		header := c.readBuf.Next(HeadLength)
		c.size = int(binary.BigEndian.Uint32(header))

		// packet length limitation
		if env.Safe && c.size > MaxPacketSize {
			return fmt.Errorf("%w, size:%d", ErrPacketSizeExcced, c.size)
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
	buf := make([]byte, 15)

	// encode flag
	flag := byte(m.Type)
	code, err := c.dictionary.IndexRoute(m.Route)
	compressed := c.compressed.Load() && err == nil
	if !compressed {
		flag |= msgRouteNotCompressMask
	}
	flag |= byte(m.Branch) << 4
	buf[offset] = flag
	offset++

	// encode version ID
	binary.BigEndian.PutUint32(buf[offset:], m.ShortVer)
	offset += 4

	// encode msg ID
	binary.BigEndian.PutUint64(buf[offset:], m.ID)
	offset += 8

	// encode route
	if compressed {
		// encode compressed route ID
		binary.BigEndian.PutUint16(buf[offset:], uint16(code))
	} else {
		rl := uint16(len(m.Route))

		// encode route string length
		binary.BigEndian.PutUint16(buf[offset:], rl)

		// encode route string
		buf = append(buf, []byte(m.Route)...)
	}

	// encode body
	var plain = make([]byte, len(m.Data)+4)
	binary.BigEndian.PutUint32(plain, uint32(len(m.Data)))

	copy(plain[4:], m.Data)

	r := len(plain) % aes.BlockSize
	if r != 0 {
		plain = append(plain, make([]byte, aes.BlockSize-r)...)
	}

	cryptData, err := aes.Encode(aesKey, plain)
	if err != nil {
		return nil, err
	}

	buf = append(buf, cryptData...)

	return buf, nil
}

func (c *CodecEntity) DecodeMessage(data []byte) (*message.Message, error) {
	if len(data) < msgHeadLength {
		return nil, ErrInvalidMessage
	}
	var offset uint64 = 0

	// decode flag
	m := message.New()
	flag := data[offset]
	offset++
	m.Type = message.Type(flag & msgTypeMask)
	m.Branch = uint32((flag & msgBranchMask) >> 4)
	c.compressed.Store(flag&msgRouteNotCompressMask == 0)
	if !m.TypeValid() {
		return nil, ErrWrongMessageType
	}

	// decode version ID
	m.ShortVer = binary.BigEndian.Uint32(data[offset:])
	offset += 4

	// decode msg ID
	m.ID = binary.BigEndian.Uint64(data[offset:])
	offset += 8

	// decode route
	if c.compressed.Load() {
		// decode compressed route ID
		code := binary.BigEndian.Uint16(data[offset:])
		route, err := c.dictionary.IndexCode(uint32(code))
		if err != nil {
			return nil, err
		}
		m.Route = route
		offset += 2
	} else {
		// decode route string length
		rl := binary.BigEndian.Uint16(data[offset:])
		offset += 2

		if offset+uint64(rl) > uint64(len(data)) {
			return nil, ErrInvalidRouteLength
		}

		// decode route string
		m.Route = string(data[offset:(offset + uint64(rl))])
		offset += uint64(rl)
	}

	// decode body
	plain, err := aes.Decode(aesKey, data[offset:])
	if err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(plain)
	m.Data = plain[4 : 4+length]

	return m, nil
}

type Codec struct {
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
