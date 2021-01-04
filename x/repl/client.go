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

	"github.com/lonng/nano/connector"
	"github.com/lonng/nano/log"
	"github.com/lonng/nano/message"
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
			connector.WithSerializer(opt.Serializer),
			connector.WithLogger(logger),
		),
		IncomingMsgChan: make(chan *message.Message, 10),
	}
}

// NewWebsocket returns a new websocket client
func NewWebsocket(path string) *Client {
	if localHandler == nil {
		initLocalHandler()
	}
	wc := &Client{
		Connector: *connector.NewConnector(
			connector.WithWSPath(path),
			connector.WithSerializer(opt.Serializer),
			connector.WithLogger(logger),
		),
		IncomingMsgChan: make(chan *message.Message, 10),
	}
	wc.Connector.OnUnexpectedEvent(UnexpectedEventCb(wc))
	return wc
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

		err = opt.Serializer.Unmarshal(push.Data, pushStruct)
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
	pc := newClient()
	// 设置服务器push过来消息的callback
	pc.Connector.OnUnexpectedEvent(UnexpectedEventCb(pc))
	return pc
}

// Connect to server
func (pc *Client) Connect(addr string) error {
	return pc.Connector.Start(addr)
}

// ConnectWS to websocket server
// func (pc *Client) ConnectWS(addr string) error {
// 	return pc.Connector.StartWS(addr)
// }

// Disconnect the client
func (pc *Client) Disconnect() {
	pc.Connector.Close()
}

// SendRequest sends a request to the server
func (pc *Client) SendRequest(route string, data []byte) (uint, error) {
	requestStruct, err := routeMessage(route)
	if err != nil {
		return 0, err
	}
	err = json.Unmarshal(data, requestStruct)
	if err != nil {
		return 0, err
	}

	err = pc.Connector.Request(route, requestStruct, func(data interface{}) {
		response := data.(*message.Message)

		responseStruct, err := routeMessage(response.Route)
		if err != nil {
			log.Println(err.Error())
			return
		}
		err = opt.Serializer.Unmarshal(response.Data, responseStruct)
		if err != nil {
			log.Printf("unmarshal error data:%v", response.Data)
			return
		}
		jsonData, err := json.Marshal(responseStruct)
		if err != nil {
			log.Printf("JSON marshal error data:%v", responseStruct)
			return
		}
		response.Data = jsonData
		pc.IncomingMsgChan <- response
	})
	if err != nil {
		return 0, err
	}
	return uint(pc.Connector.GetMid()), nil
}

// SendNotify sends a notify to the server
func (pc *Client) SendNotify(route string, data []byte) error {
	notifyStruct, err := routeMessage(route)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, notifyStruct)
	if err != nil {
		return err
	}
	return pc.Connector.Notify(route, notifyStruct)
}
