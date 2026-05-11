package ws

import (
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
	sendBufferSize = 256
)

type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	lobbyCode string
	playerID  uuid.UUID
}

func NewClient(hub *Hub, conn *websocket.Conn, lobbyCode string, playerID uuid.UUID) *Client {
	return &Client{
		hub:       hub,
		conn:      conn,
		send:      make(chan []byte, sendBufferSize),
		lobbyCode: lobbyCode,
		playerID:  playerID,
	}
}

func (c *Client) PlayerID() uuid.UUID { return c.playerID }
func (c *Client) LobbyCode() string   { return c.lobbyCode }

// ReadPump reads messages from the ws connection and routes them.
// Runs in its own goroutine. Owns reads on c.conn.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		// TODO: route incoming messages to game logic once it exists
	}
}

// WritePump drains c.send into the ws connection and sends periodic pings.
// Runs in its own goroutine. Owns writes on c.conn.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// hub closed the channel — send close frame and exit
				c.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
