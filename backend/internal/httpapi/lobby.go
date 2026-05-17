package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"who-among-you/internal/lobby"
)

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
		status := http.StatusNotFound
		if errors.Is(err, lobby.ErrGameAlreadyStarted) {
			status = http.StatusConflict
		}
		writeError(w, status, err.Error())
		return
	}

	h.broadcastLobbyState(snap)
	h.maybeStartCountdown(snap.Code, snap)

	writeJSON(w, http.StatusOK, map[string]any{
		"player": player,
		"lobby":  snap,
	})
}
