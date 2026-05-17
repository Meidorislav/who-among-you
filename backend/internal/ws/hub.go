package ws

import (
	"log"

	"github.com/google/uuid"
)

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
	direct     chan directMessage
	query      chan playerQuery
}

type Message struct {
	LobbyCode string
	Data      []byte
}

type directMessage struct {
	Client *Client
	Data   []byte
}

type playerQuery struct {
	LobbyCode string
	PlayerID  uuid.UUID
	Result    chan bool
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Message),
		direct:     make(chan directMessage),
		query:      make(chan playerQuery),
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

// SendTo delivers data to a single client. Used for state hand-offs to a
// freshly connected client without spamming everyone else.
func (h *Hub) SendTo(c *Client, data []byte) {
	h.direct <- directMessage{Client: c, Data: data}
}

// HasClientForPlayer reports whether any currently-registered WS client
// matches the given lobby+player. Used to skip leave-broadcasts when a
// reconnect (e.g. StrictMode remount, refresh) is already in flight.
func (h *Hub) HasClientForPlayer(lobbyCode string, playerID uuid.UUID) bool {
	result := make(chan bool, 1)
	h.query <- playerQuery{LobbyCode: lobbyCode, PlayerID: playerID, Result: result}
	return <-result
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
			set := h.clients[msg.LobbyCode]
			log.Printf("ws: broadcast lobby=%s recipients=%d", msg.LobbyCode, len(set))
			for client := range set {
				select {
				case client.send <- msg.Data:
				default:
					h.removeClient(client)
				}
			}

		case msg := <-h.direct:
			set := h.clients[msg.Client.lobbyCode]
			if set == nil || !set[msg.Client] {
				// Client was already unregistered — drop silently.
				continue
			}
			select {
			case msg.Client.send <- msg.Data:
			default:
				h.removeClient(msg.Client)
			}

		case q := <-h.query:
			found := false
			for c := range h.clients[q.LobbyCode] {
				if c.playerID == q.PlayerID {
					found = true
					break
				}
			}
			q.Result <- found
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
