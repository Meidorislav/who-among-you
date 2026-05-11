package httpapi

import (
	"encoding/json"
	"net/http"
	"who-among-you/internal/lobby"
	"who-among-you/internal/ws"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Handler struct {
	Lobbies *lobby.Lobbies
	Hub     *ws.Hub
}

func NewHandler(lobbies *lobby.Lobbies, hub *ws.Hub) *Handler {
	return &Handler{Lobbies: lobbies, Hub: hub}
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

	if err := h.Lobbies.JoinLobby(req.LobbyCode, player); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	event, _ := json.Marshal(map[string]any{
		"type":   "player_joined",
		"player": player,
	})
	h.Hub.Broadcast(req.LobbyCode, event)

	l, _ := h.Lobbies.GetLobby(req.LobbyCode)
	writeJSON(w, http.StatusOK, map[string]any{
		"player":  player,
		"players": l.Players,
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
// Utility function
// ---------------------------------------------------------------------------

// writeJSON insert Content-Type: application/json and encodes body in JSON.
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// writeError write JSON-object {"error": "..."} with status code.
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
