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

package serializer

import (
	"fmt"

	"github.com/aura-studio/nano/serializer/auto"
	"github.com/aura-studio/nano/serializer/json"
	"github.com/aura-studio/nano/serializer/protobuf"
	"github.com/aura-studio/nano/serializer/rawstring"
)

type (

	// Marshaler represents a marshal interface
	Marshaler interface {
		Marshal(interface{}) ([]byte, error)
	}

	// Unmarshaler represents a Unmarshal interface
	Unmarshaler interface {
		Unmarshal([]byte, interface{}) error
	}

	// Serializer is the interface that groups the basic Marshal and Unmarshal methods.
	Serializer interface {
		Marshaler
		Unmarshaler
	}
)

type SerializerType uint32

const (
	Auto SerializerType = iota
	JSON
	Protobuf
	RawString
)

var serializerTypeStrMap = map[SerializerType]string{
	Auto:      "Auto",
	JSON:      "JSON",
	Protobuf:  "Protobuf",
	RawString: "RawString",
}

var serializerTypeSerializerMap = map[SerializerType]Serializer{
	Auto:      auto.NewSerializer(),
	JSON:      json.NewSerializer(),
	Protobuf:  protobuf.NewSerializer(),
	RawString: rawstring.NewSerializer(),
}

func (t SerializerType) String() string {
	if s, ok := serializerTypeStrMap[t]; ok {
		return s
	}
	return "Unknown"
}

func (t SerializerType) Serializer() Serializer {
	if s, ok := serializerTypeSerializerMap[t]; ok {
		return s
	}
	panic(fmt.Errorf("serializer type %v not found", t))
}

func ParseSerializerType(s string) (SerializerType, error) {
	for k, v := range serializerTypeStrMap {
		if v == s {
			return k, nil
		}
	}
	return 0, fmt.Errorf("serializer type %v not found", s)
}
