package game

import (
	"math/rand"
	"sync"
	"time"
)

// Question holds localized question text.
type Question struct {
	TextEn string
	TextRu string
}

// QuestionSource produces questions for a game.
// Implementations must be safe for concurrent use.
type QuestionSource interface {
	Len() int
	Draw(count int) []Question
}

type MockQuestions struct {
	mu   sync.Mutex
	pool []Question
	rng  *rand.Rand
}

func NewMockQuestions() *MockQuestions {
	return &MockQuestions{
		pool: []Question{
			{TextEn: "Who among you sleeps the most?", TextRu: "Кто из вас больше всех спит?"},
			{TextEn: "Who among you is the best cook?", TextRu: "Кто из вас лучше всех готовит?"},
			{TextEn: "Who among you is late the most often?", TextRu: "Кто из вас чаще всех опаздывает?"},
			{TextEn: "Who among you spends the most money?", TextRu: "Кто из вас больше всех тратит деньги?"},
			{TextEn: "Who among you is the best dancer?", TextRu: "Кто из вас лучше всех танцует?"},
			{TextEn: "Who among you laughs the loudest?", TextRu: "Кто из вас громче всех смеётся?"},
			{TextEn: "Who among you travels the most?", TextRu: "Кто из вас больше всех путешествует?"},
			{TextEn: "Who among you is most afraid of heights?", TextRu: "Кто из вас больше всех боится высоты?"},
			{TextEn: "Who among you takes the longest to get ready?", TextRu: "Кто из вас дольше всех собирается?"},
			{TextEn: "Who among you drinks the most coffee?", TextRu: "Кто из вас больше всех пьёт кофе?"},
		},
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (m *MockQuestions) Len() int {
	return len(m.pool)
}

func (m *MockQuestions) Draw(count int) []Question {
	m.mu.Lock()
	defer m.mu.Unlock()

	return drawQuestions(m.pool, count, m.rng)
}

func drawQuestions(pool []Question, count int, rng *rand.Rand) []Question {
	if count <= 0 || len(pool) == 0 {
		return nil
	}

	result := make([]Question, 0, count)
	for len(result) < count {
		shuffled := append([]Question(nil), pool...)
		rng.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		remaining := count - len(result)
		if remaining > len(shuffled) {
			remaining = len(shuffled)
		}
		result = append(result, shuffled[:remaining]...)
	}
	return result
}
