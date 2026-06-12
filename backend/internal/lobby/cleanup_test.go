package lobby

import (
	"testing"
)

func TestLobbyCleanupInAllStates(t *testing.T) {
	l := InitLobbies()
	p1 := NewPlayer("P1")
	code := l.NewLobby(p1)

	// Case 1: Cleanup in Waiting state
	_, _, deleted := l.RemovePlayer(code, p1.PlayerID)
	if !deleted {
		t.Error("expected lobby to be deleted in Waiting state")
	}
	if _, exists := l.Lobbies[code]; exists {
		t.Error("lobby still exists in map after deletion")
	}

	// Case 2: Cleanup in Playing state
	p2 := NewPlayer("P2")
	code = l.NewLobby(p2)
	l.StartGame(code)
	_, _, deleted = l.RemovePlayer(code, p2.PlayerID)
	if !deleted {
		t.Error("expected lobby to be deleted in Playing state")
	}
	if _, exists := l.Lobbies[code]; exists {
		t.Error("lobby still exists in map after deletion in Playing state")
	}
}
