package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type room struct {
	// forward is channel which have messages for send another client
	forward chan []byte
	// join is channel for client which want to join chatroom
	join chan *client
	// leave is channel for client which want to leave chatroom
	leave chan *client
	// clients has all client which have been in the chatroom
	clients map[*client]bool
}

// newRoom is create chatroom
func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}

func (r *room) run() {
	for {
		select {

		case client := <-r.join:
			// join
			r.clients[client] = true

		case client := <-r.leave:
			// leave
			delete(r.clients, client)
			close(client.send)

		case msg := <-r.forward:
			// send messages to all clients
			for client := range r.clients {
				select {
				case client.send <- msg:
					// send message
				default:
					// fail to send message
					delete(r.clients, client)
					close(client.send)
				}
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	socket, err := upgrader.Upgrade(w, req, nil)

	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}

	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}

	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()

}
