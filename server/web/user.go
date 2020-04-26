package web

import (
	"github.com/gorilla/websocket"
)

type user struct {
	socket *websocket.Conn
	send   chan []byte
	room   *room
}

func (c *user) read() {
	defer c.socket.Close()
	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			return
		}

		c.room.forward <- msg
	}
}

func (c *user) write() {
	defer c.socket.Close()
	for msg := range c.send {
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}