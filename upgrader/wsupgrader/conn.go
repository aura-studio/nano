package wsupgrader

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// Conn is an adapter to t.Conn, which implements all t.Conn
// interface base on *websocket.Conn
type Conn struct {
	r      *http.Request
	conn   *websocket.Conn
	typ    int // message type
	reader io.Reader
}

// NewWSConn return an initialized *WSConn
func NewConn(r *http.Request, conn *websocket.Conn) *Conn {
	return &Conn{r: r, conn: conn}
}

// Read reads data from the connection.
// Read can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (c *Conn) Read(b []byte) (int, error) {
	if c.reader == nil {
		t, r, err := c.conn.NextReader()
		if err != nil {
			return 0, err
		}
		c.typ = t
		c.reader = r
	}

	n, err := c.reader.Read(b)
	if err != nil && err != io.EOF {
		return n, err
	} else if err == io.EOF {
		_, r, err := c.conn.NextReader()
		if err != nil {
			return 0, err
		}
		c.reader = r
	}

	return n, nil
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (c *Conn) Write(b []byte) (int, error) {
	err := c.conn.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return 0, err
	}

	return len(b), nil
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
	return newWSRemoteAddr(c.conn, c.r)
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

type wsRemoteAddr struct {
	conn *websocket.Conn
	r    *http.Request
}

func newWSRemoteAddr(conn *websocket.Conn, r *http.Request) *wsRemoteAddr {
	return &wsRemoteAddr{
		conn: conn,
		r:    r,
	}
}

func (ws *wsRemoteAddr) Network() string {
	return ws.conn.RemoteAddr().Network()
}

func (ws *wsRemoteAddr) String() string {
	XForwardFor := ws.r.Header.Get("X-Forwarded-For")
	if len(XForwardFor) == 0 {
		return ws.r.RemoteAddr
	}
	index := strings.LastIndex(ws.r.RemoteAddr, ":")
	return fmt.Sprintf("%s%s", XForwardFor, ws.r.RemoteAddr[index:])
}
