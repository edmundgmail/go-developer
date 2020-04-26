package web

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type room struct {
	forward chan []byte

	join  chan *user
	leave chan *user

	users map[*user]bool
}

func (r *room) Run() {
	for {
		select {
		case user := <-r.join:
			r.users[user] = true
		case user := <-r.leave:
			delete(r.users, user)
			close(user.send)
		case msg := <-r.forward:
			for user := range r.users {
				user.send <- msg
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}
	user := &user{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}
	r.join <- user
	defer func() { r.leave <- user }()
	go user.write()
	user.read()
}

func NewRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *user),
		leave:   make(chan *user),
		users:   make(map[*user]bool),
	}
}
