// Copyright (c) nano Authors. All Rights Reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package message

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Type represents the type of message, which could be Request/Notify/Response/Push
type Type byte

// Message types
const (
	Request  Type = 0x00
	Notify        = 0x01
	Response      = 0x02
	Push          = 0x03
)

const (
	msgRouteNotCompressMask = 0x08
	msgTypeMask             = 0x07
	msgRouteLengthMask      = 0xFF
	msgHeadLength           = 0x02
)

var types = map[Type]string{
	Request:  "Request",
	Notify:   "Notify",
	Response: "Response",
	Push:     "Push",
}

func (t Type) String() string {
	return types[t]
}

// Errors that could be occurred in message codec
var (
	ErrWrongMessageType  = errors.New("wrong message type")
	ErrInvalidMessage    = errors.New("invalid message")
	ErrRouteInfoNotFound = errors.New("route info not found in dictionary")
)

// Message represents a unmarshaled message or a message which to be marshaled
type Message struct {
	Type       Type   // message type
	ID         uint64 // unique id, zero while notify mode
	Route      string // route for locating service
	Data       []byte // payload
	compressed bool   // is message compressed
}

// New returns a new message instance
func New() *Message {
	return &Message{}
}

// String, implementation of fmt.Stringer interface
func (m *Message) String() string {
	return fmt.Sprintf("%s %s (%dbytes)", types[m.Type], m.Route, len(m.Data))
}

func invalidType(t Type) bool {
	return t < Request || t > Push

}

// Encode marshals message to binary format. Different message types is corresponding to
// different message header, message types is identified by 2-4 bit of flag field. The
// relationship between message types and message header is presented as follows:
// The figure above indicates that the bit does not affect the type of message.
// See ref: https://github.com/lonnng/nano/blob/master/docs/communication_protocol.md
func Encode(m *Message, routes map[string]uint16) ([]byte, error) {
	if invalidType(m.Type) {
		return nil, ErrWrongMessageType
	}
	var offset uint64 = 0
	buf := make([]byte, 11)

	// encode flag
	flag := byte(m.Type)
	code, compressed := routes[m.Route]
	if !compressed {
		flag |= msgRouteNotCompressMask
	}
	buf[offset] = byte(flag)
	offset++

	// encode msg ID
	binary.BigEndian.PutUint64(buf[offset:], m.ID)
	offset += 8

	// encode route
	if compressed {
		// encode compressed route ID
		binary.BigEndian.PutUint16(buf[offset:], code)
		offset += 2
	} else {
		rl := uint16(len(m.Route))

		// encode route string length
		binary.BigEndian.PutUint16(buf[offset:], rl)
		offset += 2

		// encode route string
		buf = append(buf, []byte(m.Route)...)
		offset += uint64(rl)
	}

	buf = append(buf, m.Data...)
	return buf, nil
}

// Decode unmarshal the bytes slice to a message
func Decode(data []byte, codes map[uint16]string) (*Message, bool, error) {
	if len(data) < msgHeadLength {
		return nil, false, ErrInvalidMessage
	}
	var offset uint64 = 0

	// decode flag
	m := New()
	flag := data[offset]
	offset++
	m.Type = Type(flag & msgTypeMask)
	m.compressed = flag&msgRouteNotCompressMask == 0
	if invalidType(m.Type) {
		return nil, false, ErrWrongMessageType
	}

	// decode msg ID
	m.ID = binary.BigEndian.Uint64(data[offset:])
	offset += 8

	// decode route
	if m.compressed {
		// decode compressed route ID
		code := binary.BigEndian.Uint16(data[offset:])
		route, ok := codes[code]
		if !ok {
			return nil, false, ErrRouteInfoNotFound
		}
		m.Route = route
		offset += 2
	} else {
		// decode route string length
		rl := binary.BigEndian.Uint16(data[offset:])
		offset += 2

		// decode route string
		m.Route = string(data[offset:(offset + uint64(rl))])
		offset += uint64(rl)
	}

	// decode data
	m.Data = data[offset:]
	return m, m.compressed, nil
}
