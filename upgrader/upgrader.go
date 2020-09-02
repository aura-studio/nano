package upgrader

import (
	"net"
	"net/http"
)

type Upgrader interface {
	Upgrade(w http.ResponseWriter, r *http.Request, params map[string]string) (net.Conn, error)
}
