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
	"fmt"

	"github.com/aura-studio/nano/cluster/clusterpb"
)

// Type represents the type of message, which could be Request/Notify/Response/Push
type Type byte

// Message types
const (
	Request Type = iota
	Notify
	Response
	Push
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

// Message represents a unmarshaled message or a message which to be marshaled
type Message struct {
	Type       Type   // message type
	Branch     uint32 // client branch
	VersionNum uint32 // message short version
	ID         uint64 // unique id, zero while notify mode
	Route      string // route for locating service
	Data       []byte // payload
}

// New returns a new message instance
func New() *Message {
	return &Message{}
}

// String, implementation of fmt.Stringer interface
func (m *Message) String() string {
	return fmt.Sprintf("%s %s (%dbytes)", types[m.Type], m.Route, len(m.Data))
}

func (m *Message) TypeValid() bool {
	return m.Type >= Request && m.Type <= Push
}

func (m *Message) Deserialize(v interface{}) error {
	return Serializer.Deserialize(m.Route, m.Data, v)
}

func (m *Message) Serialize(v interface{}) error {
	var err error
	m.Data, err = Serializer.Serialize(m.Route, v)
	return err
}

func Register(items []*clusterpb.MessageItem) {
	Dictionary.Register(items)
	Serializer.Register(items)
}

func Serialize(route string, v interface{}) ([]byte, error) {
	return Serializer.Serialize(route, v)
}

func Deserialize(route string, payload []byte, v interface{}) error {
	return Serializer.Deserialize(route, payload, v)
}

func Route(version uint32, route string) (uint32, error) {
	return Dictionary.Route(version, route)
}

func Code(version uint32, code uint32) (string, error) {
	return Dictionary.Code(version, code)
}
