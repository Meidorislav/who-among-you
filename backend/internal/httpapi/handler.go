package httpapi

import (
	"net/http"
	"sync"
	"time"
	"who-among-you/internal/game"
	"who-among-you/internal/lobby"
	"who-among-you/internal/ws"

	"github.com/gorilla/websocket"
)

type Handler struct {
	Lobbies *lobby.Lobbies
	Hub     *ws.Hub
	Games   *game.Manager

	countdownsMu sync.Mutex
	countdowns   map[string]*time.Timer
}

func NewHandler(lobbies *lobby.Lobbies, hub *ws.Hub, games *game.Manager) *Handler {
	return &Handler{
		Lobbies:    lobbies,
		Hub:        hub,
		Games:      games,
		countdowns: make(map[string]*time.Timer),
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // TODO: tighten for prod
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
