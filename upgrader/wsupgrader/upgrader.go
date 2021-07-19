package wsupgrader

import (
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

type Upgrader struct {
	*websocket.Upgrader
}

func NewWSUpgrader() *Upgrader {
	return &Upgrader{
		Upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(_ *http.Request) bool { return true },
		},
	}
}

func (u *Upgrader) Upgrade(w http.ResponseWriter, r *http.Request, params map[string]string) (net.Conn, error) {
	conn, err := u.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return NewConn(conn), nil
}
