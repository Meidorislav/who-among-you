package game

import (
	"math/rand"
	"sync"
	"time"
)

// Question holds localized question text.
type Question struct {
	TextEn   string
	TextRu   string
	Category string
}

// QuestionSource produces questions for a game.
// Implementations must be safe for concurrent use.
// An empty categories slice means "all categories".
type QuestionSource interface {
	Len(categories []string) int
	Draw(count int, categories []string) []Question
	Categories() []string
}

func distinctCategories(pool []Question) []string {
	seen := make(map[string]bool)
	var result []string
	for _, q := range pool {
		if !seen[q.Category] {
			seen[q.Category] = true
			result = append(result, q.Category)
		}
	}
	return result
}

func filterPool(pool []Question, categories []string) []Question {
	if len(categories) == 0 {
		return pool
	}
	set := make(map[string]bool, len(categories))
	for _, c := range categories {
		set[c] = true
	}
	filtered := make([]Question, 0, len(pool))
	for _, q := range pool {
		if set[q.Category] {
			filtered = append(filtered, q)
		}
	}
	return filtered
}

type MockQuestions struct {
	mu   sync.Mutex
	pool []Question
	rng  *rand.Rand
}

func NewMockQuestions() *MockQuestions {
	return &MockQuestions{
		pool: []Question{
			{TextEn: "Who among you sleeps the most?", TextRu: "Кто из вас больше всех спит?", Category: "habits"},
			{TextEn: "Who among you is the best cook?", TextRu: "Кто из вас лучше всех готовит?", Category: "food"},
			{TextEn: "Who among you is late the most often?", TextRu: "Кто из вас чаще всех опаздывает?", Category: "habits"},
			{TextEn: "Who among you spends the most money?", TextRu: "Кто из вас больше всех тратит деньги?", Category: "money"},
			{TextEn: "Who among you is the best dancer?", TextRu: "Кто из вас лучше всех танцует?", Category: "skills"},
			{TextEn: "Who among you laughs the loudest?", TextRu: "Кто из вас громче всех смеётся?", Category: "personality"},
			{TextEn: "Who among you travels the most?", TextRu: "Кто из вас больше всех путешествует?", Category: "travel"},
			{TextEn: "Who among you is most afraid of heights?", TextRu: "Кто из вас больше всех боится высоты?", Category: "fears"},
			{TextEn: "Who among you takes the longest to get ready?", TextRu: "Кто из вас дольше всех собирается?", Category: "habits"},
			{TextEn: "Who among you drinks the most coffee?", TextRu: "Кто из вас больше всех пьёт кофе?", Category: "food"},
		},
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (m *MockQuestions) Categories() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return distinctCategories(m.pool)
}

func (m *MockQuestions) Len(categories []string) int {
	return len(filterPool(m.pool, categories))
}

func (m *MockQuestions) Draw(count int, categories []string) []Question {
	m.mu.Lock()
	defer m.mu.Unlock()

	return drawQuestions(filterPool(m.pool, categories), count, m.rng)
}

func drawQuestions(pool []Question, count int, rng *rand.Rand) []Question {
	if count <= 0 || len(pool) == 0 {
		return nil
	}
	if count > len(pool) {
		count = len(pool)
	}
	shuffled := append([]Question(nil), pool...)
	rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled[:count]
}
