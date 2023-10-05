package initializers

import (
	"github.com/gofiber/websocket/v2"
)

var Clients = make(map[string]*websocket.Conn)
