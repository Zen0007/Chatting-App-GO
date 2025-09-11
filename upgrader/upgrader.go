// Package upgrader
package upgrader

import (
	"net/http"
	"github.com/gorilla/websocket"
)

type Upgrade struct {
	Upgrader websocket.Upgrader
}

func Upgrader() *Upgrade {
	return &Upgrade{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}
