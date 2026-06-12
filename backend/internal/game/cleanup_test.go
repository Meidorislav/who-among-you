package game

import (
	"testing"
	"who-among-you/internal/lobby"

	"github.com/google/uuid"
)

type mockBroadcaster struct{}
func (m *mockBroadcaster) Broadcast(code string, data []byte) {}

func TestGameCleanup(t *testing.T) {
	qs := NewMockQuestions()
	mb := &mockBroadcaster{}
	m := NewManager(qs, mb)
	
	code := "TEST01"
	players := []uuid.UUID{uuid.New(), uuid.New()}
	
	m.Start(code, players, lobby.Settings{QuestionCount: 5})
	
	if _, ok := m.games[code]; !ok {
		t.Fatal("game was not started")
	}
	
	m.DeleteGame(code)
	
	if _, ok := m.games[code]; ok {
		t.Error("game still exists after DeleteGame")
	}
}
