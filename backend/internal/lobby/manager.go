package lobby

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
)

type LobbyStatus string

const (
	StatusWaiting  LobbyStatus = "waiting"
	StatusPlaying  LobbyStatus = "playing"
	StatusFinished LobbyStatus = "finished"
)

const minPlayersToStart = 2

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
		return Snapshot{}, errors.New("lobby not found")
	}
	if lobby.Status != StatusWaiting {
		return Snapshot{}, errors.New("game already started")
	}
	lobby.Players = append(lobby.Players, player)
	return snapshot(lobby), nil
}

// SetReady toggles the ready flag for a player. If all players are ready and
// minPlayersToStart is reached, the lobby transitions to playing atomically.
// Returns the new snapshot, whether the game just started, and ok=false if
// the lobby or player doesn't exist.
func (l *Lobbies) SetReady(code string, playerID uuid.UUID, ready bool) (snap Snapshot, gameStarted bool, ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	lobby, exists := l.Lobbies[code]
	if !exists {
		return Snapshot{}, false, false
	}
	if lobby.Status != StatusWaiting {
		return snapshot(lobby), false, true
	}

	found := false
	for i := range lobby.Players {
		if lobby.Players[i].PlayerID == playerID {
			lobby.Players[i].Ready = ready
			found = true
			break
		}
	}
	if !found {
		return Snapshot{}, false, false
	}

	if len(lobby.Players) >= minPlayersToStart && allReady(lobby.Players) {
		lobby.Status = StatusPlaying
		gameStarted = true
	}

	return snapshot(lobby), gameStarted, true
}

func allReady(players []Player) bool {
	for _, p := range players {
		if !p.Ready {
			return false
		}
	}
	return true
}
