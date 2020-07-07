package connector

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/lonng/nano/cluster"
	"github.com/lonng/nano/log"
	"github.com/lonng/nano/serialize/protobuf"

	"github.com/lonng/nano/codec"
	"github.com/lonng/nano/message"
	"github.com/lonng/nano/packet"
)

type (

	// Callback represents the callback type which will be called
	// when the correspond events is occurred.
	Callback func(data interface{})

	// Connector is a tiny Nano client
	Connector struct {
		Options

		conn      net.Conn       // low-level connection
		codec     *codec.Decoder // decoder
		die       chan struct{}  // connector close channel
		chSend    chan []byte    // send queue
		mid       uint64         // message id
		connected int32          // connected state 1: disconnected : 0

		// events handler
		muEvents        sync.RWMutex
		events          map[string]Callback // stores all events by key:event name value:callback
		unexpectedEvent Callback            // un registered event run this callback

		// response handler
		muResponses sync.RWMutex
		responses   map[uint64]Callback

		connectedCallback func() // connected callback

		routes map[string]uint16 // copy system routes for agent
		codes  map[uint16]string // copy system codes for agent
	}
)

// NewConnector create a new Connector
func NewConnector(opts ...Option) *Connector {
	c := &Connector{
		Options: Options{
			dictionary: make(map[string]uint16),
			serializer: protobuf.NewSerializer(),
		},
		die:       make(chan struct{}),
		codec:     codec.NewDecoder(),
		chSend:    make(chan []byte, 256),
		mid:       1,
		connected: 0,
		events:    map[string]Callback{},
		responses: map[uint64]Callback{},
		routes:    make(map[string]uint16),
		codes:     make(map[uint16]string),
	}

	for i := range opts {
		opt := opts[i]
		opt(&c.Options)
	}

	if c.Options.logger != nil {
		log.SetLogger(c.Options.logger)
	}

	c.routes, c.codes = message.ParseDictionary(c.dictionary)
	return c
}

// Start connects to the server and send/recv between the c/s
func (c *Connector) Start(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	c.conn = conn

	go c.write()

	go c.read()

	atomic.StoreInt32(&c.connected, 1)
	go c.connectedCallback()

	return nil
}

// StartWS connects to websocket server
func (c *Connector) StartWS(addr string) error {
	u := url.URL{Scheme: "ws", Host: addr, Path: c.wsPath}
	dialer := websocket.DefaultDialer
	var conn *websocket.Conn
	var err error
	conn, _, err = dialer.Dial(u.String(), nil)
	if err != nil {
		u.Scheme = "wss"
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		conn, _, err = dialer.Dial(u.String(), nil)
		if err != nil {
			return err
		}
	}

	c.conn, err = cluster.NewWSConn(conn)
	if err != nil {
		return err
	}

	go c.write()

	go c.read()

	atomic.StoreInt32(&c.connected, 1)
	go c.connectedCallback()

	return nil
}

// Name returns the name for connector
func (c *Connector) Name() string {
	return c.name
}

// OnConnected set the callback which will be called when the client connected to the server
func (c *Connector) OnConnected(callback func()) {
	c.connectedCallback = callback
}

// GetMid returns current message id
func (c *Connector) GetMid() uint64 {
	return c.mid
}

// Request send a request to server and register a callbck for the response
func (c *Connector) Request(route string, v interface{}, callback Callback) error {
	data, err := c.Serialize(v)
	if err != nil {
		return err
	}

	msg := &message.Message{
		Type:  message.Request,
		Route: route,
		ID:    c.mid,
		Data:  data,
	}

	c.setResponseHandler(c.mid, callback)
	if err := c.sendMessage(msg); err != nil {
		c.setResponseHandler(c.mid, nil)
		return err
	}

	return nil
}

// RawRequest send a request []byte to server and register a callbck for the response
func (c *Connector) RawRequest(route string, data []byte, callback Callback) error {
	msg := &message.Message{
		Type:  message.Request,
		Route: route,
		ID:    c.mid,
		Data:  data,
	}

	c.setResponseHandler(c.mid, callback)
	if err := c.sendMessage(msg); err != nil {
		c.setResponseHandler(c.mid, nil)
		return err
	}

	return nil
}

// Notify send a notification to server
func (c *Connector) Notify(route string, v interface{}) error {
	data, err := c.Serialize(v)
	if err != nil {
		return err
	}
	msg := &message.Message{
		Type:  message.Notify,
		Route: route,
		Data:  data,
	}
	return c.sendMessage(msg)
}

// RawNotify send a []byte notification to server
func (c *Connector) RawNotify(route string, data []byte) error {
	msg := &message.Message{
		Type:  message.Notify,
		Route: route,
		Data:  data,
	}
	return c.sendMessage(msg)
}

// On add the callback for the event
func (c *Connector) On(event string, callback Callback) {
	c.muEvents.Lock()
	defer c.muEvents.Unlock()

	c.events[event] = callback
}

// OnUnexpectedEvent sets callback for events that are not "On"
func (c *Connector) OnUnexpectedEvent(callback Callback) {
	c.unexpectedEvent = callback
}

// Close closes the connection, and shutdown the benchmark
func (c *Connector) Close() {
	defer func() {
		recover()
	}()
	c.conn.Close()
	close(c.die)
}

// Connected returns the status whether connector is conncected
func (c *Connector) Connected() bool {
	return atomic.LoadInt32(&c.connected) == 1
}

// Serialize marshals customized data into byte slice
func (c *Connector) Serialize(v interface{}) ([]byte, error) {
	if data, ok := v.([]byte); ok {
		return data, nil
	}

	if c.serializer == nil {
		return nil, fmt.Errorf("Serializer is not set")
	}

	data, err := c.serializer.Marshal(v)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Deserialize Unmarshals byte slice into customized data
func (c *Connector) Deserialize(data []byte, v interface{}) error {
	var err error
	if c.serializer == nil {
		v = data
	}

	if c.serializer == nil {
		return fmt.Errorf("Serializer is not set")
	}

	if err = c.serializer.Unmarshal(data, v); err != nil {
		return err
	}

	return nil
}

func (c *Connector) eventHandler(event string) (Callback, bool) {
	c.muEvents.RLock()
	defer c.muEvents.RUnlock()

	cb, ok := c.events[event]
	return cb, ok
}

func (c *Connector) responseHandler(mid uint64) (Callback, bool) {
	c.muResponses.RLock()
	defer c.muResponses.RUnlock()

	cb, ok := c.responses[mid]
	return cb, ok
}

func (c *Connector) setResponseHandler(mid uint64, cb Callback) {
	c.muResponses.Lock()
	defer c.muResponses.Unlock()

	if cb == nil {
		delete(c.responses, mid)
	} else {
		c.responses[mid] = cb
	}
}

func (c *Connector) sendMessage(msg *message.Message) error {
	data, err := message.Encode(msg, c.routes)
	if err != nil {
		return err
	}

	payload, err := codec.Encode(data)
	if err != nil {
		return err
	}

	c.mid++
	c.send(payload)

	return nil
}

func (c *Connector) write() {
	defer close(c.chSend)

	for {
		select {
		case data := <-c.chSend:
			if _, err := c.conn.Write(data); err != nil {
				log.Println(err)
				c.Close()
			}

		case <-c.die:
			atomic.StoreInt32(&c.connected, 0)
			return
		}
	}
}

func (c *Connector) send(data []byte) {
	c.chSend <- data
}

func (c *Connector) read() {
	buf := make([]byte, 2048)

	for {
		n, err := c.conn.Read(buf)
		if err != nil {
			log.Println(err)
			c.Close()
			return
		}

		packets, err := c.codec.Decode(buf[:n])
		if err != nil {
			log.Println(err)
			c.Close()
			return
		}

		for i := range packets {
			p := packets[i]
			c.processPacket(p)
		}
	}
}

func (c *Connector) processPacket(p *packet.Packet) {
	msg, _, err := message.Decode(p.Data, c.codes)
	if err != nil {
		log.Println(err)
		return
	}
	c.processMessage(msg)
}

func (c *Connector) processMessage(msg *message.Message) {
	switch msg.Type {
	case message.Push:
		cb, ok := c.eventHandler(msg.Route)
		if ok {
			cb(msg)
		} else {
			c.unexpectedEvent(msg)
		}

	case message.Response:
		cb, ok := c.responseHandler(msg.ID)
		if !ok {
			log.Println("response handler not found", msg.ID)
			return
		}

		cb(msg)
		c.setResponseHandler(msg.ID, nil)
	}
}
