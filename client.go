package main

import (
	"github.com/gorilla/websocket"
)

// client is one user which doing chat
type client struct {
	// socket is websocket for this client
	socket *websocket.Conn
	// send is channel which send messages
	send chan []byte
	// room is chatroom which join this client
	room *room
}

func (c *client) read() {
	for {
		if _, msg, err := c.socket.ReadMessage(); err == nil {
			c.room.forward <- msg
		} else {
			break
		}
	}
	c.socket.Close()
}

func (c *client) write() {
	for msg := range c.send {
		if err := c.socket.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
	c.socket.Close()
}
