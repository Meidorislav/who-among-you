package game

import (
	"math/rand"
	"sync"
	"time"
)

// QuestionSource produces the next question for a round.
// Implementations must be safe for concurrent use.
type QuestionSource interface {
	Next() string
}

type MockQuestions struct {
	mu   sync.Mutex
	pool []string
	rng  *rand.Rand
}

func NewMockQuestions() *MockQuestions {
	return &MockQuestions{
		pool: []string{
			"Who among you sleeps the most?",
			"Who among you is the best cook?",
			"Who among you is late the most often?",
			"Who among you spends the most money?",
			"Who among you is the best dancer?",
			"Who among you laughs the loudest?",
			"Who among you travels the most?",
			"Who among you is most afraid of heights?",
			"Who among you takes the longest to get ready?",
			"Who among you drinks the most coffee?",
		},
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (m *MockQuestions) Next() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.pool[m.rng.Intn(len(m.pool))]
}
