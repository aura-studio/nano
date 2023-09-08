package httpupgrader

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/aura-studio/nano/log"
	"github.com/aura-studio/nano/message"
	"github.com/aura-studio/nano/options"
	"github.com/aura-studio/nano/packet"
	"github.com/aura-studio/nano/serializer"

	"github.com/aura-studio/nano/codec"
	"github.com/aura-studio/nano/env"
)

// Conn is an adapter to t.Conn, which implements all t.Conn
// interface base on *websocket.Conn
type Conn struct {
	w           http.ResponseWriter
	r           *http.Request
	conn        net.Conn
	brw         *bufio.ReadWriter
	params      map[string]string
	codecEntity codec.CodecEntity
	readBuf     io.Reader
	readDone    atomic.Bool
	writeDone   atomic.Bool
}

// NewConn return an initialized *WSConn
func NewConn(w http.ResponseWriter, r *http.Request, conn net.Conn, brw *bufio.ReadWriter, params map[string]string) *Conn {
	return &Conn{
		w:           w,
		r:           r,
		conn:        conn,
		brw:         brw,
		params:      params,
		codecEntity: options.Default.Codec.Entity(nil),
		readBuf:     nil,
	}
}

// Read reads data from the connection.
// Read can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (c *Conn) Read(b []byte) (int, error) {
	if c.readDone.Load() {
		if c.writeDone.Load() {
			return 0, io.EOF
		}
		time.Sleep(time.Minute)
		return 0, io.EOF
	}

	if c.readBuf == nil {
		var (
			err      error
			data     []byte
			route    string
			dataMap  = make(map[string]interface{})
			dataType uint32
		)

		route = c.params["route"]
		query := c.r.URL.Query()
		if len(data) == 0 {
			for k, v := range query {
				dataMap[k] = v[0]
			}
		}

		contentType := strings.Split(c.r.Header.Get("Content-Type"), ";")[0]
		switch contentType {
		case "multipart/form-data", "application/x-www-form-urlencoded":
			_ = c.r.FormValue("")
			for k, v := range c.r.Form {
				dataMap[k] = v[0]
			}
			data, err = json.Marshal(dataMap)
			if err != nil {
				return 0, err
			}
		default:
			var bodyData []byte
			if c.r.ContentLength < 0 {
				return 0, fmt.Errorf("content length is: %d", c.r.ContentLength)
			}

			bodyData = make([]byte, int(c.r.ContentLength))
			_, err := io.ReadFull(c.brw, bodyData)
			if err != nil {
				return 0, err
			}

			jsonMap := make(map[string]interface{})
			if err := json.Unmarshal(bodyData, &jsonMap); err == nil {
				for k, v := range jsonMap {
					dataMap[k] = v
				}
			}

			if len(bodyData) > 0 { // binary
				data = bodyData
				dataType = serializer.JSONType
			} else { // json
				data, err = json.Marshal(dataMap)
				if err != nil {
					return 0, err
				}
				switch c.r.Header.Get("Content-Type") {
				case "application/json":
					dataType = serializer.JSONType
				case "application/x-protobuf":
					dataType = serializer.ProtobufType
				default:
					dataType = serializer.AutoType
				}
			}
		}

		if env.Debug {
			log.Infof("http: Type=Request, Route=%s, Len=%d, DataType=%s", route, len(data), serializer.TypeToStr(dataType))
		}

		msg := &message.Message{
			Type:      message.Request,
			Route:     route,
			ID:        1,
			UnixTime:  uint32(time.Now().Unix()),
			SessionID: 0, // 0 is ok for pure http
			DataType:  dataType,
			Data:      data,
		}
		m, err := c.codecEntity.EncodeMessage(msg)
		if err != nil {
			return 0, err
		}
		packets := []*packet.Packet{{Length: len(m), Data: m}}
		p, err := c.codecEntity.EncodePacket(packets)
		if err != nil {
			return 0, err
		}

		buf := new(bytes.Buffer)
		_, err = buf.Write(p)
		if err != nil {
			return 0, err
		}
		c.readBuf = buf
	}

	n, err := c.readBuf.Read(b)
	if err == io.EOF {
		c.readDone.Store(true)
		return n, nil
	}

	return n, err
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (c *Conn) Write(b []byte) (int, error) {
	defer c.writeDone.Store(true)

	packets, err := c.codecEntity.DecodePacket(b)
	if err != nil {
		return 0, err
	}
	if len(packets) != 1 {
		return 0, fmt.Errorf("http: Error number of packets to write")
	}
	m, err := c.codecEntity.DecodeMessage(packets[0].Data)
	if err != nil {
		return 0, err
	}

	if env.Debug {
		log.Infof("http: Type=Response, Route=%s, Len=%d, Data=%+v", m.Route, len(m.Data), string(m.Data))
	}

	header := fmt.Sprintf("HTTP/1.0 200 OK\r\nContent-Length: %d\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n", len(m.Data))
	nHeader, err := c.brw.Write([]byte(header))
	if err != nil {
		return 0, err
	}
	nBody, err := c.brw.Write(m.Data)
	if err != nil {
		return 0, err
	}
	err = c.brw.Flush()
	if err != nil {
		return 0, err
	}
	n := nHeader + nBody

	return n, nil
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (c *Conn) Close() error {
	return c.conn.Close()
}

// LocalAddr returns the local network address.
func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	return newHTTPRemoteAddr(c.conn, c.r)
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to Read or
// Write. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (c *Conn) SetDeadline(t time.Time) error {
	if err := c.conn.SetReadDeadline(t); err != nil {
		return err
	}

	return c.conn.SetWriteDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

type httpRemoteAddr struct {
	conn net.Conn
	r    *http.Request
}

func newHTTPRemoteAddr(conn net.Conn, r *http.Request) *httpRemoteAddr {
	return &httpRemoteAddr{
		conn: conn,
		r:    r,
	}
}

func (hra *httpRemoteAddr) Network() string {
	return hra.conn.RemoteAddr().Network()
}

func (hra *httpRemoteAddr) String() string {
	XForwardFor := hra.r.Header.Get("X-Forwarded-For")
	if len(XForwardFor) == 0 {
		return hra.r.RemoteAddr
	}
	index := strings.LastIndex(hra.r.RemoteAddr, ":")
	return fmt.Sprintf("%s%s", XForwardFor, hra.r.RemoteAddr[index:])
}
