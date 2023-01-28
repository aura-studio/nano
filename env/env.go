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

// env represents the environment of the current process, includes
// work path and config path etc.
package env

import (
	"time"

	"github.com/aura-studio/nano/serializer"
	"github.com/aura-studio/snowflake"
	"google.golang.org/grpc"
)

var (
	// Wd is working path
	Wd string

	// Die waits for end application
	Die chan bool

	// ConnDie waits for application die
	ConnDie chan bool

	// Debug enables Debug mode
	Debug bool

	// SafeMode enable Safe mode
	Safe bool

	// TimerPrecision indicates the precision of timer, default is time.Second
	TimerPrecision = time.Second

	// Serializer is message serializer, json or protobuf
	SerializerType serializer.SerializerType

	// GrpcOptions is options for grpc
	GrpcOptions = []grpc.DialOption{grpc.WithInsecure()}

	// Version
	Version string

	// ShortVersion is short for Version
	ShortVersion uint32

	// SnowflakeNode is snowflake node
	SnowflakeNode *snowflake.Node
)

func init() {
	Die = make(chan bool)
	ConnDie = make(chan bool)
	Debug = false
	Safe = true
	SerializerType = serializer.Auto
}
