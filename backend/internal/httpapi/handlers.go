package httpapi

import (
	"encoding/json"
	"net/http"
	"who-among-you/internal/game"
	"who-among-you/internal/lobby"
	"who-among-you/internal/ws"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Handler struct {
	Lobbies *lobby.Lobbies
	Hub     *ws.Hub
	Games   *game.Manager
}

func NewHandler(lobbies *lobby.Lobbies, hub *ws.Hub, games *game.Manager) *Handler {
	return &Handler{Lobbies: lobbies, Hub: hub, Games: games}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // TODO: tighten for prod
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func (h *Handler) CreateLobby(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Nickname string `json:"nickname"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Nickname == "" {
		writeError(w, http.StatusBadRequest, "nickname required")
		return
	}

	player := lobby.NewPlayer(req.Nickname)
	code := h.Lobbies.NewLobby(player)

	writeJSON(w, http.StatusCreated, map[string]any{
		"code":   code,
		"player": player,
	})
}

func (h *Handler) JoinLobby(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Nickname  string `json:"nickname"`
		LobbyCode string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Nickname == "" {
		writeError(w, http.StatusBadRequest, "nickname required")
		return
	}

	player := lobby.NewPlayer(req.Nickname)

	snap, err := h.Lobbies.JoinLobby(req.LobbyCode, player)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	h.broadcastLobbyState(snap)

	writeJSON(w, http.StatusOK, map[string]any{
		"player": player,
		"lobby":  snap,
	})
}

func (h *Handler) WS(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	playerIDStr := r.URL.Query().Get("player_id")

	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid player_id")
		return
	}

	if !h.Lobbies.HasPlayer(code, playerID) {
		writeError(w, http.StatusForbidden, "player not in lobby")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return // Upgrade already wrote an error response
	}

	client := ws.NewClient(h.Hub, conn, code, playerID)
	h.Hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}

// ---------------------------------------------------------------------------
// WS message routing
// ---------------------------------------------------------------------------

// HandleMessage implements ws.MessageHandler. Runs in the ReadPump goroutine
// of the originating client.
func (h *Handler) HandleMessage(playerID uuid.UUID, lobbyCode string, data []byte) {
	var env struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return
	}

	switch env.Type {
	case "set_ready":
		var msg struct {
			Ready bool `json:"ready"`
		}
		if err := json.Unmarshal(data, &msg); err != nil {
			return
		}
		h.handleSetReady(lobbyCode, playerID, msg.Ready)

	case "vote":
		var msg struct {
			TargetPlayerID uuid.UUID `json:"target_player_id"`
		}
		if err := json.Unmarshal(data, &msg); err != nil {
			return
		}
		h.Games.Vote(lobbyCode, playerID, msg.TargetPlayerID)
	}
}

func (h *Handler) handleSetReady(code string, playerID uuid.UUID, ready bool) {
	snap, gameStarted, ok := h.Lobbies.SetReady(code, playerID, ready)
	if !ok {
		return
	}
	h.broadcastLobbyState(snap)
	if gameStarted {
		h.broadcastEvent(code, map[string]any{"type": "game_started"})
		h.Games.Start(code, playerIDs(snap.Players))
	}
}

func playerIDs(players []lobby.Player) []uuid.UUID {
	ids := make([]uuid.UUID, len(players))
	for i, p := range players {
		ids[i] = p.PlayerID
	}
	return ids
}

func (h *Handler) broadcastLobbyState(snap lobby.Snapshot) {
	h.broadcastEvent(snap.Code, map[string]any{
		"type":  "lobby_state",
		"lobby": snap,
	})
}

func (h *Handler) broadcastEvent(code string, payload map[string]any) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	h.Hub.Broadcast(code, data)
}

// ---------------------------------------------------------------------------
// HTTP utilities
// ---------------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
