package game

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	RoundDuration   = 45 * time.Second
	ResultsDuration = 5 * time.Second
	TotalRounds     = 10
)

type Phase string

const (
	PhaseVoting   Phase = "voting"
	PhaseResults  Phase = "results"
	PhaseFinished Phase = "finished"
)

// Broadcaster dispatches a serialized event to all clients of a lobby.
type Broadcaster interface {
	Broadcast(lobbyCode string, data []byte)
}

type Manager struct {
	mu          sync.Mutex
	games       map[string]*Game
	questions   QuestionSource
	broadcaster Broadcaster
}

func NewManager(questions QuestionSource, broadcaster Broadcaster) *Manager {
	return &Manager{
		games:       make(map[string]*Game),
		questions:   questions,
		broadcaster: broadcaster,
	}
}

// Start creates a game for the lobby and triggers the first round.
// No-op if a game already exists for this lobby.
func (m *Manager) Start(lobbyCode string, players []uuid.UUID) {
	m.mu.Lock()
	if _, exists := m.games[lobbyCode]; exists {
		m.mu.Unlock()
		return
	}
	g := newGame(lobbyCode, players, m.questions, m.broadcaster)
	m.games[lobbyCode] = g
	m.mu.Unlock()

	g.startNextRound()
}

// Vote records a vote from voter for target in the current round.
// Silently dropped if the game/round is not accepting votes.
func (m *Manager) Vote(lobbyCode string, voter, target uuid.UUID) {
	m.mu.Lock()
	g, ok := m.games[lobbyCode]
	m.mu.Unlock()
	if !ok {
		return
	}
	g.handleVote(voter, target)
}

// ---------------------------------------------------------------------------
// Game (per-lobby state)
// ---------------------------------------------------------------------------

type Game struct {
	mu sync.Mutex

	lobbyCode    string
	players      []uuid.UUID
	currentRound int
	phase        Phase
	question     string
	deadline     time.Time
	votes        map[uuid.UUID]uuid.UUID // voter -> target
	scores       map[uuid.UUID]int
	timer        *time.Timer

	questions   QuestionSource
	broadcaster Broadcaster
}

func newGame(code string, players []uuid.UUID, qs QuestionSource, b Broadcaster) *Game {
	g := &Game{
		lobbyCode:   code,
		players:     append([]uuid.UUID(nil), players...),
		scores:      make(map[uuid.UUID]int, len(players)),
		questions:   qs,
		broadcaster: b,
	}
	for _, p := range players {
		g.scores[p] = 0
	}
	return g
}

func (g *Game) startNextRound() {
	g.mu.Lock()
	g.currentRound++

	if g.currentRound > TotalRounds {
		g.phase = PhaseFinished
		g.broadcastLocked(map[string]any{
			"type":   "game_finished",
			"scores": g.scores,
		})
		g.mu.Unlock()
		return
	}

	g.phase = PhaseVoting
	g.question = g.questions.Next()
	g.votes = make(map[uuid.UUID]uuid.UUID, len(g.players))
	g.deadline = time.Now().Add(RoundDuration)

	roundN := g.currentRound
	g.timer = time.AfterFunc(RoundDuration, func() {
		g.endRoundIfStill(roundN)
	})

	g.broadcastLocked(map[string]any{
		"type":     "round_started",
		"round":    g.currentRound,
		"total":    TotalRounds,
		"question": g.question,
		"deadline": g.deadline.Unix(),
		"players":  g.players,
	})
	g.mu.Unlock()
}

func (g *Game) handleVote(voter, target uuid.UUID) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.phase != PhaseVoting {
		return
	}
	if !g.hasPlayer(voter) || !g.hasPlayer(target) {
		return
	}

	g.votes[voter] = target

	if len(g.votes) == len(g.players) {
		g.endRoundLocked()
	}
}

// endRoundIfStill is called from the round timer. No-op if the round has
// already been ended (e.g., everyone voted early).
func (g *Game) endRoundIfStill(round int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.currentRound != round || g.phase != PhaseVoting {
		return
	}
	g.endRoundLocked()
}

func (g *Game) endRoundLocked() {
	if g.timer != nil {
		g.timer.Stop()
	}

	counts := make(map[uuid.UUID]int, len(g.players))
	for _, target := range g.votes {
		counts[target]++
	}

	maxCount := 0
	for _, c := range counts {
		if c > maxCount {
			maxCount = c
		}
	}

	winners := make([]uuid.UUID, 0)
	if maxCount > 0 {
		for target, c := range counts {
			if c == maxCount {
				winners = append(winners, target)
				g.scores[target]++
			}
		}
	}

	g.phase = PhaseResults
	g.broadcastLocked(map[string]any{
		"type":    "round_ended",
		"round":   g.currentRound,
		"votes":   counts,
		"scores":  g.scores,
		"winners": winners,
	})

	roundN := g.currentRound
	g.timer = time.AfterFunc(ResultsDuration, func() {
		g.advanceFromResults(roundN)
	})
}

func (g *Game) advanceFromResults(round int) {
	g.mu.Lock()
	if g.currentRound != round || g.phase != PhaseResults {
		g.mu.Unlock()
		return
	}
	g.mu.Unlock()
	g.startNextRound()
}

func (g *Game) hasPlayer(id uuid.UUID) bool {
	for _, p := range g.players {
		if p == id {
			return true
		}
	}
	return false
}

// broadcastLocked serializes the event and pushes it to the broadcaster.
// Caller MUST hold g.mu (the map fields are read during marshaling).
func (g *Game) broadcastLocked(event map[string]any) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	g.broadcaster.Broadcast(g.lobbyCode, data)
}
