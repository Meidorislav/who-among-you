package httpapi

import (
	"encoding/json"
	"net/http"
	"who-among-you/internal/lobby"
)

type Handler struct {
	Lobbies *lobby.Lobbies
}

func NewHandler(lobbies *lobby.Lobbies) *Handler {
	return &Handler{Lobbies: lobbies}
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

	lobby, _ := h.Lobbies.GetLobby(req.LobbyCode)
	writeJSON(w, http.StatusOK, map[string]any{
		"players": lobby.Players,
	})
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

// mustJSON serializes the value in JSON. Error panic
func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic("json marshal: " + err.Error())
	}
	b = append(b, '\n')
	return b
}
