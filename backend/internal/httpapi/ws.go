package httpapi

import (
	"encoding/json"
	"net/http"
	"who-among-you/internal/lobby"
	"who-among-you/internal/ws"

	"github.com/google/uuid"
)

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
