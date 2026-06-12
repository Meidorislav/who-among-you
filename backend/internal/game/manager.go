package game

import (
	"encoding/json"
	"sync"
	"time"
	"who-among-you/internal/lobby"
	"who-among-you/internal/ws"

	"github.com/google/uuid"
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

type Options struct {
	TotalRounds   int
	RoundDuration time.Duration
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
func (m *Manager) Start(lobbyCode string, players []uuid.UUID, settings lobby.Settings) {
	m.mu.Lock()
	if _, exists := m.games[lobbyCode]; exists {
		m.mu.Unlock()
		return
	}
	g := newGame(lobbyCode, players, m.options(settings), m.questions, m.broadcaster)
	m.games[lobbyCode] = g
	m.mu.Unlock()

	g.startNextRound()
}

func (m *Manager) options(settings lobby.Settings) Options {
	totalRounds := settings.QuestionCount
	if totalRounds == lobby.AllQuestions {
		totalRounds = m.questions.Len()
	}
	if totalRounds <= 0 {
		totalRounds = lobby.DefaultQuestionCount
	}

	var roundDuration time.Duration
	if settings.RoundDurationSeconds > lobby.NoRoundTimeLimit {
		roundDuration = time.Duration(settings.RoundDurationSeconds) * time.Second
	}

	return Options{TotalRounds: totalRounds, RoundDuration: roundDuration}
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

func (m *Manager) ReadyForNextRound(lobbyCode string, player uuid.UUID) {
	m.mu.Lock()
	g, ok := m.games[lobbyCode]
	m.mu.Unlock()
	if !ok {
		return
	}
	g.readyForNextRound(player)
}

func (m *Manager) DeleteGame(lobbyCode string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if g, ok := m.games[lobbyCode]; ok {
		g.stop()
		delete(m.games, lobbyCode)
	}
}

func (m *Manager) SendCurrentRound(lobbyCode string, broadcaster Broadcaster, client interface{}) {
	m.mu.Lock()
	g, ok := m.games[lobbyCode]
	m.mu.Unlock()
	if !ok {
		return
	}
	g.sendCurrentRound(broadcaster, client)
}

// ---------------------------------------------------------------------------
// Game (per-lobby state)
// ---------------------------------------------------------------------------

type Game struct {
	mu sync.Mutex

	lobbyCode    string
	players      []uuid.UUID
	options      Options
	questions    []Question
	currentRound int
	phase        Phase
	question     Question
	deadline     time.Time
	votes        map[uuid.UUID]uuid.UUID // voter -> target
	nextReady    map[uuid.UUID]bool
	scores       map[uuid.UUID]int
	timer        *time.Timer

	broadcaster Broadcaster
}

func (g *Game) stop() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.timer != nil {
		g.timer.Stop()
	}
}

func newGame(code string, players []uuid.UUID, options Options, qs QuestionSource, b Broadcaster) *Game {
	g := &Game{
		lobbyCode:   code,
		players:     append([]uuid.UUID(nil), players...),
		options:     options,
		questions:   qs.Draw(options.TotalRounds),
		scores:      make(map[uuid.UUID]int, len(players)),
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

	if g.currentRound > g.options.TotalRounds {
		g.phase = PhaseFinished
		g.broadcastLocked(map[string]any{
			"type":   "game_finished",
			"scores": g.scores,
		})
		g.mu.Unlock()
		return
	}

	g.phase = PhaseVoting
	g.question = g.questions[g.currentRound-1]
	g.votes = make(map[uuid.UUID]uuid.UUID, len(g.players))
	g.nextReady = nil
	if g.options.RoundDuration > 0 {
		g.deadline = time.Now().Add(g.options.RoundDuration)
	} else {
		g.deadline = time.Time{}
	}

	roundN := g.currentRound
	if g.options.RoundDuration > 0 {
		g.timer = time.AfterFunc(g.options.RoundDuration, func() {
			g.endRoundIfStill(roundN)
		})
	} else {
		g.timer = nil
	}
	deadline := int64(0)
	if !g.deadline.IsZero() {
		deadline = g.deadline.Unix()
	}

	g.broadcastLocked(map[string]any{
		"type":                   "round_started",
		"round":                  g.currentRound,
		"total":                  g.options.TotalRounds,
		"question_en":            g.question.TextEn,
		"question_ru":            g.question.TextRu,
		"deadline":               deadline,
		"round_duration_seconds": int(g.options.RoundDuration / time.Second),
		"players":                g.players,
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
	g.nextReady = make(map[uuid.UUID]bool, len(g.players))
	g.broadcastLocked(map[string]any{
		"type":       "round_ended",
		"round":      g.currentRound,
		"votes":      counts,
		"scores":     g.scores,
		"winners":    winners,
		"next_ready": readyIDs(g.nextReady),
	})
}

func (g *Game) readyForNextRound(player uuid.UUID) {
	startNext := false

	g.mu.Lock()
	if g.phase != PhaseResults || !g.hasPlayer(player) {
		g.mu.Unlock()
		return
	}

	if g.nextReady[player] {
		delete(g.nextReady, player)
	} else {
		g.nextReady[player] = true
		if len(g.nextReady) == len(g.players) {
			startNext = true
		}
	}

	g.broadcastLocked(map[string]any{
		"type":       "next_round_state",
		"round":      g.currentRound,
		"next_ready": readyIDs(g.nextReady),
	})
	g.mu.Unlock()

	if startNext {
		g.startNextRound()
	}
}

func (g *Game) sendCurrentRound(broadcaster Broadcaster, client interface{}) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.phase == PhaseFinished {
		return
	}

	deadline := int64(0)
	if !g.deadline.IsZero() {
		deadline = g.deadline.Unix()
	}

	// Always send round_started first
	event := map[string]any{
		"type":                   "round_started",
		"round":                  g.currentRound,
		"total":                  g.options.TotalRounds,
		"question_en":            g.question.TextEn,
		"question_ru":            g.question.TextRu,
		"deadline":               deadline,
		"round_duration_seconds": int(g.options.RoundDuration / time.Second),
		"players":                g.players,
	}

	if data, err := json.Marshal(event); err == nil {
		if hub, ok := broadcaster.(*ws.Hub); ok {
			if c, ok := client.(*ws.Client); ok {
				hub.SendTo(c, data)
			}
		}
	}

	// If in results phase, also send round_ended
	if g.phase == PhaseResults {
		resultEvent := map[string]any{
			"type":       "round_ended",
			"round":      g.currentRound,
			"votes":      g.votes,
			"scores":     g.scores,
			"winners":    g.getWinners(),
			"next_ready": readyIDs(g.nextReady),
		}

		if data, err := json.Marshal(resultEvent); err == nil {
			if hub, ok := broadcaster.(*ws.Hub); ok {
				if c, ok := client.(*ws.Client); ok {
					hub.SendTo(c, data)
				}
			}
		}
	}
}

func (g *Game) getWinners() []uuid.UUID {
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
			}
		}
	}
	return winners
}

func readyIDs(ready map[uuid.UUID]bool) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(ready))
	for id := range ready {
		ids = append(ids, id)
	}
	return ids
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
