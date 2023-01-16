// Copyright (c) TFG Co. All Rights Reserved.
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

package repl

import (
	"encoding/json"

	"github.com/aura-studio/nano/log"
	"github.com/aura-studio/nano/serializer"

	"github.com/aura-studio/nano/connector"
	"github.com/aura-studio/nano/message"
)

// Client struct
type Client struct {
	connector.Connector
	IncomingMsgChan chan *message.Message
}

// ConnectedStatus returns the the current connection status
func (pc *Client) ConnectedStatus() bool {
	return pc.Connector.Connected()
}

// MsgChannel return the incoming message channel
func (pc *Client) MsgChannel() chan *message.Message {
	return pc.IncomingMsgChan
}

// Return the basic structure for the Client struct.
func newClient() *Client {
	return &Client{
		Connector: *connector.NewConnector(
			connector.WithIsWebSocket(options.IsWebSocket),
			connector.WithCodec(options.Codec),
			connector.WithDictionary(options.Dictionary),
			connector.WithSerializerType(options.SerializerType),
			connector.WithLogger(logger),
			connector.WithWSPath(options.WSPath),
			connector.WithBranch(options.Branch),
		),
		IncomingMsgChan: make(chan *message.Message, 10),
	}
}

// UnexpectedEventCb returns a function to deal with un listened event
func UnexpectedEventCb(pc *Client) func(data interface{}) {
	return func(data interface{}) {
		push := data.(*message.Message)
		pushStruct, err := routeMessage(push.Route)
		if err != nil {
			log.Println(err.Error())
			return
		}

		err = options.SerializerType.Serializer().Unmarshal(push.Data, pushStruct)
		if err != nil {
			log.Printf("unmarshal error data:%v ", push.Data)
			return
		}

		jsonData, err := json.Marshal(pushStruct)
		if err != nil {
			log.Printf("JSON marshal error data:%v", pushStruct)
			return
		}
		push.Data = jsonData
		pc.IncomingMsgChan <- push
	}
}

// NewClient returns a new client with the auto documentation route.
func NewClient() *Client {
	if localHandler == nil {
		initLocalHandler()
	}
	c := newClient()
	// 设置服务器push过来消息的callback
	if options.EnableUnexpected {
		c.Connector.OnUnexpectedEvent(UnexpectedEventCb(c))
	}

	return c
}

// Connect to server
func (pc *Client) Connect(addr string) error {
	return pc.Connector.Start(addr)
}

// Disconnect the client
func (pc *Client) Disconnect() {
	pc.Connector.Close()
}

// SendRequest sends a request to the server
func (pc *Client) SendRequest(route string, data []byte) (uint, error) {
	var request interface{}
	if options.SerializerType != serializer.JSON {
		v, err := routeMessage(route)
		if err != nil {
			return 0, err
		}
		switch v.(type) {
		case []byte:
			request = data
		default:
			err = json.Unmarshal(data, v)
			if err != nil {
				return 0, err
			}
			request = v
		}
	} else {
		request = data
	}

	err := pc.Connector.Request(route, request, func(data interface{}) {
		response := data.(*message.Message)
		if options.SerializerType != serializer.JSON {
			v, err := routeMessage(response.Route)
			if err != nil {
				log.Println(err.Error())
				return
			}
			switch v.(type) {
			case []byte:
			default:
				err = options.SerializerType.Serializer().Unmarshal(response.Data, v)
				if err != nil {
					log.Printf("unmarshal error data:%v", response.Data)
					return
				}
				response.Data, err = json.Marshal(v)
				if err != nil {
					log.Printf("JSON marshal error data:%v", v)
					return
				}
			}
		}
		pc.IncomingMsgChan <- response
	})
	if err != nil {
		return 0, err
	}
	return uint(pc.Connector.GetMid()), nil
}

// SendNotify sends a notify to the server
func (pc *Client) SendNotify(route string, data []byte) error {
	var notify interface{}
	if options.SerializerType != serializer.JSON {
		v, err := routeMessage(route)
		if err != nil {
			return err
		}
		switch v.(type) {
		case []byte:
			notify = data
		default:
			err = json.Unmarshal(data, v)
			if err != nil {
				return err
			}
			notify = v
		}
	} else {
		notify = data
	}
	return pc.Connector.Notify(route, notify)
}
