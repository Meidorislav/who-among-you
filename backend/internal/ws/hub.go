package ws

import "github.com/google/uuid"

// MessageHandler routes inbound ws messages from clients into application logic.
// Implementations must be safe for concurrent use — ReadPump calls it directly.
type MessageHandler interface {
	HandleMessage(playerID uuid.UUID, lobbyCode string, data []byte)
}

type Hub struct {
	clients map[string]map[*Client]bool // lobbyCode -> set of clients
	handler MessageHandler

	register   chan *Client
	unregister chan *Client
	broadcast  chan Message
}

type Message struct {
	LobbyCode string
	Data      []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Message),
	}
}

// SetHandler installs the message handler. Must be called before Run starts
// receiving traffic. Not safe to call after clients are connected.
func (h *Hub) SetHandler(handler MessageHandler) { h.handler = handler }

func (h *Hub) Register(c *Client)   { h.register <- c }
func (h *Hub) Unregister(c *Client) { h.unregister <- c }
func (h *Hub) Broadcast(lobbyCode string, data []byte) {
	h.broadcast <- Message{LobbyCode: lobbyCode, Data: data}
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
			h.removeClient(client)

		case msg := <-h.broadcast:
			for client := range h.clients[msg.LobbyCode] {
				select {
				case client.send <- msg.Data:
				default:
					h.removeClient(client)
				}
			}
		}
	}
}

func (h *Hub) removeClient(c *Client) {
	set, ok := h.clients[c.lobbyCode]
	if !ok {
		return
	}
	if _, exists := set[c]; !exists {
		return
	}
	delete(set, c)
	close(c.send)
	if len(set) == 0 {
		delete(h.clients, c.lobbyCode)
	}
}
