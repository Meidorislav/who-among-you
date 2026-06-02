package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	"who-among-you/internal/lobby"
	"who-among-you/internal/ws"

	"github.com/google/uuid"
)

// leaveGracePeriod buys time for a quick reconnect (React StrictMode double-
// mount, full page refresh, brief network hiccup) before we declare a player
// gone and drop them from the lobby. Dev refreshes with Vite HMR cold-start
// can take a couple of seconds, so we err on the side of patience.
const leaveGracePeriod = 30 * time.Second

// startCountdownDuration is the pre-game window during which any player can
// unready (or a new player can join) to abort the start. The frontend animates
// a ring over this duration.
const startCountdownDuration = 5 * time.Second

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

	if snap, ok := h.Lobbies.GetLobby(code); ok {
		if data, err := json.Marshal(map[string]any{
			"type":  "lobby_state",
			"lobby": snap,
		}); err == nil {
			h.Hub.SendTo(client, data)
		}
	}

	h.Games.SendCurrentRound(code, h.Hub, client)

	log.Printf("ws: player %s connected to lobby %s", playerID, code)
	client.ReadPump() // blocks until the connection drops
	log.Printf("ws: player %s disconnected from lobby %s", playerID, code)

	h.scheduleLeave(code, playerID)
}

// scheduleLeave drops the player from the lobby after a grace period, unless
// another WS connection for the same player has appeared in the meantime
// (e.g. StrictMode remount, refresh, reconnect).
func (h *Handler) scheduleLeave(code string, playerID uuid.UUID) {
	time.Sleep(leaveGracePeriod)
	if h.Hub.HasClientForPlayer(code, playerID) {
		return
	}
	snap, removed, deleted := h.Lobbies.RemovePlayer(code, playerID)
	if !removed {
		return
	}
	if deleted {
		h.cancelCountdown(code)
		return
	}
	h.broadcastLobbyState(snap)
	h.maybeStartCountdown(code, snap)
	log.Printf("ws: player %s removed from lobby %s after grace", playerID, code)
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
		if !h.Lobbies.HasPlayer(lobbyCode, playerID) {
			return
		}
		var msg struct {
			Ready bool `json:"ready"`
		}
		if err := json.Unmarshal(data, &msg); err != nil {
			return
		}
		h.handleSetReady(lobbyCode, playerID, msg.Ready)

	case "update_settings":
		if !h.Lobbies.HasPlayer(lobbyCode, playerID) {
			return
		}
		var msg struct {
			QuestionCount        int `json:"question_count"`
			RoundDurationSeconds int `json:"round_duration_seconds"`
		}
		if err := json.Unmarshal(data, &msg); err != nil {
			return
		}
		h.handleUpdateSettings(lobbyCode, playerID, lobby.Settings{
			QuestionCount:        msg.QuestionCount,
			RoundDurationSeconds: msg.RoundDurationSeconds,
		})

	case "kick_player":
		if !h.Lobbies.HasPlayer(lobbyCode, playerID) {
			return
		}
		var msg struct {
			TargetPlayerID uuid.UUID `json:"target_player_id"`
		}
		if err := json.Unmarshal(data, &msg); err != nil {
			return
		}
		h.handleKickPlayer(lobbyCode, playerID, msg.TargetPlayerID)

	case "vote":
		if !h.Lobbies.HasPlayer(lobbyCode, playerID) {
			return
		}
		var msg struct {
			TargetPlayerID uuid.UUID `json:"target_player_id"`
		}
		if err := json.Unmarshal(data, &msg); err != nil {
			return
		}
		h.Games.Vote(lobbyCode, playerID, msg.TargetPlayerID)

	case "next_round":
		if !h.Lobbies.HasPlayer(lobbyCode, playerID) {
			return
		}
		h.Games.ReadyForNextRound(lobbyCode, playerID)
	}
}

func (h *Handler) handleSetReady(code string, playerID uuid.UUID, ready bool) {
	snap, ok := h.Lobbies.SetReady(code, playerID, ready)
	if !ok {
		return
	}
	h.broadcastLobbyState(snap)
	h.maybeStartCountdown(code, snap)
}

func (h *Handler) handleUpdateSettings(code string, hostID uuid.UUID, settings lobby.Settings) {
	snap, err := h.Lobbies.UpdateSettings(code, hostID, settings)
	if err != nil {
		return
	}
	h.broadcastLobbyState(snap)
	h.maybeStartCountdown(code, snap)
}

func (h *Handler) handleKickPlayer(code string, hostID, targetID uuid.UUID) {
	snap, removed, err := h.Lobbies.KickPlayer(code, hostID, targetID)
	if err != nil || !removed {
		return
	}
	h.broadcastLobbyState(snap)
	h.maybeStartCountdown(code, snap)
	log.Printf("ws: host %s kicked player %s from lobby %s", hostID, targetID, code)
}

// maybeStartCountdown is the single decision point that turns "lobby state
// changed" into "should we be counting down to game start?". Called after
// every mutation that could flip the answer: set_ready, join, leave.
func (h *Handler) maybeStartCountdown(code string, snap lobby.Snapshot) {
	if isReadyToStart(snap.Players) {
		h.startCountdown(code)
	} else {
		h.cancelCountdown(code)
	}
}

func isReadyToStart(players []lobby.Player) bool {
	if len(players) < lobby.MinPlayersToStart {
		return false
	}
	for _, p := range players {
		if !p.Ready {
			return false
		}
	}
	return true
}

// startCountdown begins (or leaves running) the pre-start timer for a lobby.
// Broadcasts countdown_started with the absolute deadline so clients can
// animate consistently regardless of clock skew within the same browser.
func (h *Handler) startCountdown(code string) {
	h.countdownsMu.Lock()
	if _, exists := h.countdowns[code]; exists {
		h.countdownsMu.Unlock()
		return // already running; don't reset the deadline
	}
	deadline := time.Now().Add(startCountdownDuration)
	timer := time.AfterFunc(startCountdownDuration, func() {
		h.fireCountdown(code)
	})
	h.countdowns[code] = timer
	h.countdownsMu.Unlock()

	h.broadcastEvent(code, map[string]any{
		"type":     "countdown_started",
		"deadline": deadline.UnixMilli(),
	})
	log.Printf("ws: countdown started for lobby %s", code)
}

// cancelCountdown stops a pending start timer if one is active. Safe no-op
// when nothing is running. Broadcasts countdown_cancelled so the UI can
// drop back to the regular lobby view.
func (h *Handler) cancelCountdown(code string) {
	h.countdownsMu.Lock()
	timer, exists := h.countdowns[code]
	if !exists {
		h.countdownsMu.Unlock()
		return
	}
	delete(h.countdowns, code)
	h.countdownsMu.Unlock()

	timer.Stop()
	h.broadcastEvent(code, map[string]any{"type": "countdown_cancelled"})
	log.Printf("ws: countdown cancelled for lobby %s", code)
}

// fireCountdown is the AfterFunc callback. It checks the map under lock to
// avoid racing with a concurrent cancelCountdown: if the entry is gone, the
// cancel won and we abort instead of starting the game.
func (h *Handler) fireCountdown(code string) {
	h.countdownsMu.Lock()
	_, present := h.countdowns[code]
	delete(h.countdowns, code)
	h.countdownsMu.Unlock()
	if !present {
		return
	}

	snap, ok := h.Lobbies.StartGame(code)
	if !ok {
		return
	}
	h.broadcastLobbyState(snap)
	h.broadcastEvent(code, map[string]any{"type": "game_started"})
	h.Games.Start(code, playerIDs(snap.Players), snap.Settings)
	log.Printf("ws: game started for lobby %s", code)
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
