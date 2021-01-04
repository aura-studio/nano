package upgrader

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/lonng/nano/codec"
	"github.com/lonng/nano/env"
	"github.com/lonng/nano/log"
	"github.com/lonng/nano/message"
)

// httpConn is an adapter to t.Conn, which implements all t.Conn
// interface base on *websocket.Conn
type httpConn struct {
	w       http.ResponseWriter
	r       *http.Request
	conn    net.Conn
	brw     *bufio.ReadWriter
	params  map[string]string
	decoder *codec.Decoder
	readBuf io.Reader
	readEOF bool
}

// newHttpConn return an initialized *WSConn
func newHttpConn(w http.ResponseWriter, r *http.Request, conn net.Conn, brw *bufio.ReadWriter, params map[string]string) *httpConn {
	return &httpConn{
		w:       w,
		r:       r,
		conn:    conn,
		brw:     brw,
		params:  params,
		decoder: codec.NewDecoder(),
		readBuf: nil,
	}
}

// Read reads data from the connection.
// Read can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (c *httpConn) Read(b []byte) (int, error) {
	if c.readEOF {
		return c.brw.Read(b)
	}

	if c.readBuf == nil {
		var data []byte
		if c.r.ContentLength > 0 {
			data = make([]byte, c.r.ContentLength)
			n, err := c.brw.Read(data)
			if err != nil {
				return 0, err
			}
			if n != len(data) {
				return 0, fmt.Errorf("Http: body length does not equal to content length")
			}
		} else {
			data = []byte(c.r.URL.Query().Get("data"))
		}

		var route string
		if c.params["route"] != "" {
			route = c.params["route"]
		} else {
			route = c.r.URL.Query().Get("route")
		}

		if env.Debug {
			log.Infof("Http: Type=Request, Route=%s, Len=%d, Data=%+v", route, len(data), string(data))
		}

		msg := &message.Message{
			Type:  message.Request,
			Route: route,
			ID:    1,
			Data:  data,
		}
		packet, err := message.Encode(msg, nil)
		if err != nil {
			return 0, err
		}
		b, err := codec.Encode(packet)
		if err != nil {
			return 0, err
		}
		buf := new(bytes.Buffer)
		_, err = buf.Write(b)
		if err != nil {
			return 0, err
		}
		c.readBuf = buf
	}

	n, err := c.readBuf.Read(b)
	if err == io.EOF {
		c.readEOF = true
		return n, nil
	}

	return n, err
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (c *httpConn) Write(b []byte) (int, error) {
	packets, err := c.decoder.Decode(b)
	if err != nil {
		return 0, err
	}
	if len(packets) != 1 {
		return 0, fmt.Errorf("Http: Error number of packets to write")
	}
	m, _, err := message.Decode(packets[0].Data, nil)
	if err != nil {
		return 0, err
	}

	if env.Debug {
		log.Infof("Http: Type=Response, Route=%s, Len=%d, Data=%+v", m.Route, len(m.Data), string(m.Data))
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
func (c *httpConn) Close() error {
	return c.conn.Close()
}

// LocalAddr returns the local network address.
func (c *httpConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (c *httpConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
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
func (c *httpConn) SetDeadline(t time.Time) error {
	if err := c.conn.SetReadDeadline(t); err != nil {
		return err
	}

	return c.conn.SetWriteDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (c *httpConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (c *httpConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
