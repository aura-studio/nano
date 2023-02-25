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

type Type = uint32

const (
	AutoType Type = iota
	JSONType
	ProtobufType
	RawStringType
)

const (
	AutoString      = "Auto"
	JSONString      = "JSON"
	ProtobufString  = "Protobuf"
	RawStringString = "RawString"
)

func FromString(s string) Serializer {
	switch s {
	case AutoString:
		return auto.Serializer
	case JSONString:
		return json.Serializer
	case ProtobufString:
		return protobuf.Serializer
	case RawStringString:
		return rawstring.Serializer
	}
	panic(fmt.Errorf("serializer: unknown serializer %s", s))
}

func ToString(s Serializer) string {
	switch s {
	case auto.Serializer:
		return AutoString
	case json.Serializer:
		return JSONString
	case protobuf.Serializer:
		return ProtobufString
	case rawstring.Serializer:
		return RawStringString
	}
	panic(fmt.Errorf("serializer: unknown serializer %v", s))
}

func FromType(t Type) Serializer {
	switch t {
	case AutoType:
		return auto.Serializer
	case JSONType:
		return json.Serializer
	case ProtobufType:
		return protobuf.Serializer
	case RawStringType:
		return rawstring.Serializer
	}
	panic(fmt.Errorf("serializer: unknown serializer %d", t))
}

func ToType(s Serializer) Type {
	switch s {
	case auto.Serializer:
		return AutoType
	case json.Serializer:
		return JSONType
	case protobuf.Serializer:
		return ProtobufType
	case rawstring.Serializer:
		return RawStringType
	}
	panic(fmt.Errorf("serializer: unknown serializer %v", s))
}

func StrToType(s string) Type {
	switch s {
	case AutoString:
		return AutoType
	case JSONString:
		return JSONType
	case ProtobufString:
		return ProtobufType
	case RawStringString:
		return RawStringType
	}
	panic(fmt.Errorf("serializer: unknown serializer %s", s))
}

func TypeToStr(t Type) string {
	switch t {
	case AutoType:
		return AutoString
	case JSONType:
		return JSONString
	case ProtobufType:
		return ProtobufString
	case RawStringType:
		return RawStringString
	}
	panic(fmt.Errorf("serializer: unknown serializer %d", t))
}
