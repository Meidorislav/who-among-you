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
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidSettings    = errors.New("invalid settings")
)

type LobbyStatus string

const (
	StatusWaiting  LobbyStatus = "waiting"
	StatusPlaying  LobbyStatus = "playing"
	StatusFinished LobbyStatus = "finished"
)

// MinPlayersToStart is the smallest lobby size that can begin a game.
const MinPlayersToStart = 2

const (
	DefaultQuestionCount        = 10
	DefaultRoundDurationSeconds = 45
	AllQuestions                = 0
	NoRoundTimeLimit            = 0
	MaxQuestionCount            = 500
	MinRoundDurationSeconds     = 10
	MaxRoundDurationSeconds     = 300
)

type Player struct {
	Nickname string    `json:"nickname"`
	PlayerID uuid.UUID `json:"player_id"`
	Ready    bool      `json:"ready"`
}

type Settings struct {
	QuestionCount        int      `json:"question_count"`
	RoundDurationSeconds int      `json:"round_duration_seconds"`
	Categories           []string `json:"categories"`
}

type Lobby struct {
	Code     string      `json:"code"`
	Status   LobbyStatus `json:"status"`
	HostID   uuid.UUID   `json:"host_id"`
	Settings Settings    `json:"settings"`
	Players  []Player    `json:"players"`
}

type Lobbies struct {
	mu      sync.RWMutex
	Lobbies map[string]*Lobby
}

// Snapshot is an immutable view of a lobby, safe to read after the lock is released.
type Snapshot struct {
	Code     string      `json:"code"`
	Status   LobbyStatus `json:"status"`
	HostID   uuid.UUID   `json:"host_id"`
	Settings Settings    `json:"settings"`
	Players  []Player    `json:"players"`
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
	return Snapshot{
		Code:     lobby.Code,
		Status:   lobby.Status,
		HostID:   lobby.HostID,
		Settings: lobby.Settings,
		Players:  players,
	}
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
		Code:   code,
		Status: StatusWaiting,
		HostID: player.PlayerID,
		Settings: Settings{
			QuestionCount:        DefaultQuestionCount,
			RoundDurationSeconds: DefaultRoundDurationSeconds,
			Categories:           []string{},
		},
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

func (l *Lobbies) UpdateSettings(code string, hostID uuid.UUID, settings Settings) (Snapshot, error) {
	if settings.QuestionCount < AllQuestions || settings.RoundDurationSeconds < NoRoundTimeLimit {
		return Snapshot{}, ErrInvalidSettings
	}
	if settings.QuestionCount > MaxQuestionCount {
		return Snapshot{}, ErrInvalidSettings
	}
	if settings.RoundDurationSeconds > NoRoundTimeLimit &&
		(settings.RoundDurationSeconds < MinRoundDurationSeconds || settings.RoundDurationSeconds > MaxRoundDurationSeconds) {
		return Snapshot{}, ErrInvalidSettings
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	lobby, exists := l.Lobbies[code]
	if !exists {
		return Snapshot{}, ErrLobbyNotFound
	}
	if lobby.Status != StatusWaiting {
		return snapshot(lobby), ErrGameAlreadyStarted
	}
	if lobby.HostID != hostID {
		return snapshot(lobby), ErrForbidden
	}

	lobby.Settings = settings
	return snapshot(lobby), nil
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

// RemovePlayer drops a player from a lobby. Returns the new snapshot,
// whether anything actually changed, and whether the lobby itself was deleted
// (because it became empty).
func (l *Lobbies) RemovePlayer(code string, playerID uuid.UUID) (snap Snapshot, removed bool, deleted bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	lobby, exists := l.Lobbies[code]
	if !exists {
		return Snapshot{}, false, false
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
	if lobby.HostID == playerID {
		lobby.HostID = lobby.Players[0].PlayerID
	}

	return snapshot(lobby), true, false
}

func (l *Lobbies) KickPlayer(code string, hostID, targetID uuid.UUID) (Snapshot, bool, error) {
	if hostID == targetID {
		return Snapshot{}, false, ErrForbidden
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	lobby, exists := l.Lobbies[code]
	if !exists {
		return Snapshot{}, false, ErrLobbyNotFound
	}
	if lobby.Status != StatusWaiting {
		return snapshot(lobby), false, ErrGameAlreadyStarted
	}
	if lobby.HostID != hostID {
		return snapshot(lobby), false, ErrForbidden
	}

	idx := -1
	for i := range lobby.Players {
		if lobby.Players[i].PlayerID == targetID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return snapshot(lobby), false, nil
	}

	lobby.Players = append(lobby.Players[:idx], lobby.Players[idx+1:]...)
	return snapshot(lobby), true, nil
}
