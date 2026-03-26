package ws

import "github.com/gorilla/websocket"

type Client struct {
	conn      *websocket.Conn
	send      chan []byte
	lobbyCode string
	playerID  string
}

type Hub struct {
	clients map[string]map[*Client]bool // lobbyCode -> set of clients

	register   chan *Client
	unregister chan *Client
	broadcast  chan Message
}

type Message struct {
	lobbyCode string
	data      []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Message),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			if h.clients[client.lobbyCode] == nil {
				h.clients[client.lobbyCode] = make(map[*Client]bool)
			}
			h.clients[client.lobbyCode][client] = true

		case client := <-h.unregister:
			delete(h.clients[client.lobbyCode], client)
			close(client.send)

		case msg := <-h.broadcast:
			for client := range h.clients[msg.lobbyCode] {
				client.send <- msg.data
			}
		}
	}
}
