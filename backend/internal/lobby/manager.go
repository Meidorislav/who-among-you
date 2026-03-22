package lobby

import (
	"errors"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type Player struct {
	Nickname string
	PlayerID uuid.UUID
}

type Lobby struct {
	Code    string
	Players []Player
}

type Lobbies struct {
	Lobbies map[string]*Lobby
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

func (l *Lobbies) GetLobbyCode() string {
	for {
		code := generateCode()
		if _, exists := l.Lobbies[code]; !exists {
			return code
		}
	}
}

func (l *Lobbies) NewLobby(player Player) string {
	code := l.GetLobbyCode()
	lobby := &Lobby{
		Code:    code,
		Players: []Player{player},
	}
	l.Lobbies[code] = lobby
	return code
}

func (l *Lobbies) JoinLobby(code string, player Player) error {
	lobby, exists := l.Lobbies[code]
	if !exists {
		return errors.New("lobby not found")
	}
	lobby.Players = append(lobby.Players, player)
	return nil
}
