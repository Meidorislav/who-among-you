package lobby

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrLobbyNotFound      = errors.New("lobby not found")
	ErrGameAlreadyStarted = errors.New("game already started")
)

type LobbyStatus string

const (
	StatusWaiting  LobbyStatus = "waiting"
	StatusPlaying  LobbyStatus = "playing"
	StatusFinished LobbyStatus = "finished"
)

// MinPlayersToStart is the smallest lobby size that can begin a game.
const MinPlayersToStart = 2

type Player struct {
	Nickname string    `json:"nickname"`
	PlayerID uuid.UUID `json:"player_id"`
	Ready    bool      `json:"ready"`
}

type Lobby struct {
	Code    string      `json:"code"`
	Status  LobbyStatus `json:"status"`
	Players []Player    `json:"players"`
}

type Lobbies struct {
	mu      sync.RWMutex
	Lobbies map[string]*Lobby
}

// Snapshot is an immutable view of a lobby, safe to read after the lock is released.
type Snapshot struct {
	Code    string      `json:"code"`
	Status  LobbyStatus `json:"status"`
	Players []Player    `json:"players"`
}

func NewPlayer(nickname string) Player {
	return Player{
		Nickname: nickname,
		PlayerID: uuid.New(),
	}
}

func InitLobbies() *Lobbies {
	return &Lobbies{
		Lobbies: make(map[string]*Lobby),
	}
}

const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func generateCode() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 6)
	for i := range b {
		b[i] = alphabet[r.Intn(len(alphabet))]
	}
	return string(b)
}

// getLobbyCode returns an unused code. Caller MUST hold l.mu (write lock).
func (l *Lobbies) getLobbyCode() string {
	for {
		code := generateCode()
		if _, exists := l.Lobbies[code]; !exists {
			return code
		}
	}
}

// snapshot copies the lobby fields into a Snapshot. Caller MUST hold l.mu (any lock).
func snapshot(lobby *Lobby) Snapshot {
	players := make([]Player, len(lobby.Players))
	copy(players, lobby.Players)
	return Snapshot{Code: lobby.Code, Status: lobby.Status, Players: players}
}

func (l *Lobbies) HasPlayer(code string, playerID uuid.UUID) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	lobby, ok := l.Lobbies[code]
	if !ok {
		return false
	}
	for _, p := range lobby.Players {
		if p.PlayerID == playerID {
			return true
		}
	}
	return false
}

func (l *Lobbies) NewLobby(player Player) string {
	l.mu.Lock()
	defer l.mu.Unlock()

	code := l.getLobbyCode()
	lobby := &Lobby{
		Code:    code,
		Status:  StatusWaiting,
		Players: []Player{player},
	}
	l.Lobbies[code] = lobby
	return code
}

func (l *Lobbies) GetLobby(code string) (Snapshot, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	lobby, ok := l.Lobbies[code]
	if !ok {
		return Snapshot{}, false
	}
	return snapshot(lobby), true
}

func (l *Lobbies) JoinLobby(code string, player Player) (Snapshot, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	lobby, exists := l.Lobbies[code]
	if !exists {
		return Snapshot{}, ErrLobbyNotFound
	}
	if lobby.Status != StatusWaiting {
		return Snapshot{}, ErrGameAlreadyStarted
	}
	lobby.Players = append(lobby.Players, player)
	return snapshot(lobby), nil
}

// SetReady toggles the ready flag for a player. Returns the new snapshot and
// ok=false if the lobby or player doesn't exist. Does NOT transition the lobby
// status — that decision now lives in the handler, which uses a delayed
// countdown before actually starting the game.
func (l *Lobbies) SetReady(code string, playerID uuid.UUID, ready bool) (Snapshot, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	lobby, exists := l.Lobbies[code]
	if !exists {
		return Snapshot{}, false
	}
	if lobby.Status != StatusWaiting {
		return snapshot(lobby), true
	}

	for i := range lobby.Players {
		if lobby.Players[i].PlayerID == playerID {
			lobby.Players[i].Ready = ready
			return snapshot(lobby), true
		}
	}
	return Snapshot{}, false
}

// StartGame transitions a waiting lobby to playing. Returns ok=false if the
// lobby is missing or already past waiting. Called by the countdown timer after
// the pre-start delay has elapsed.
func (l *Lobbies) StartGame(code string) (Snapshot, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	lobby, exists := l.Lobbies[code]
	if !exists || lobby.Status != StatusWaiting {
		return Snapshot{}, false
	}
	lobby.Status = StatusPlaying
	return snapshot(lobby), true
}

// RemovePlayer drops a player from a waiting lobby. Returns the new snapshot,
// whether anything actually changed, and whether the lobby itself was deleted
// (because it became empty). No-op if the lobby is already playing/finished —
// disconnect during a game keeps the player record so game state stays valid.
func (l *Lobbies) RemovePlayer(code string, playerID uuid.UUID) (snap Snapshot, removed bool, deleted bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	lobby, exists := l.Lobbies[code]
	if !exists {
		return Snapshot{}, false, false
	}
	if lobby.Status != StatusWaiting {
		return snapshot(lobby), false, false
	}

	idx := -1
	for i := range lobby.Players {
		if lobby.Players[i].PlayerID == playerID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return snapshot(lobby), false, false
	}

	lobby.Players = append(lobby.Players[:idx], lobby.Players[idx+1:]...)

	if len(lobby.Players) == 0 {
		delete(l.Lobbies, code)
		return Snapshot{Code: code, Status: lobby.Status, Players: nil}, true, true
	}

	return snapshot(lobby), true, false
}
