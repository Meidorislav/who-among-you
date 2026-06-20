package database

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
	"who-among-you/internal/game"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresQuestions struct {
	mu   sync.Mutex
	pool []game.Question
	rng  *rand.Rand
}

func NewPostgresQuestions(ctx context.Context, db *pgxpool.Pool) (*PostgresQuestions, error) {
	rows, err := db.Query(ctx, "SELECT text_en, text_ru, category FROM questions")
	if err != nil {
		return nil, fmt.Errorf("query questions: %w", err)
	}
	defer rows.Close()

	var questions []game.Question
	for rows.Next() {
		var q game.Question
		if err := rows.Scan(&q.TextEn, &q.TextRu, &q.Category); err != nil {
			return nil, fmt.Errorf("scan question: %w", err)
		}
		questions = append(questions, q)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}
	if len(questions) == 0 {
		return nil, fmt.Errorf("no questions found in database")
	}

	return &PostgresQuestions{
		pool: questions,
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

func (p *PostgresQuestions) Categories() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	seen := make(map[string]bool)
	var result []string
	for _, q := range p.pool {
		if !seen[q.Category] {
			seen[q.Category] = true
			result = append(result, q.Category)
		}
	}
	return result
}

func (p *PostgresQuestions) Len(categories []string) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(filterPool(p.pool, categories))
}

func (p *PostgresQuestions) Draw(count int, categories []string) []game.Question {
	p.mu.Lock()
	defer p.mu.Unlock()

	pool := filterPool(p.pool, categories)
	if count <= 0 || len(pool) == 0 {
		return nil
	}

	result := make([]game.Question, 0, count)
	for len(result) < count {
		shuffled := append([]game.Question(nil), pool...)
		p.rng.Shuffle(len(shuffled), func(i, j int) {
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

func filterPool(pool []game.Question, categories []string) []game.Question {
	if len(categories) == 0 {
		return pool
	}
	set := make(map[string]bool, len(categories))
	for _, c := range categories {
		set[c] = true
	}
	filtered := make([]game.Question, 0, len(pool))
	for _, q := range pool {
		if set[q.Category] {
			filtered = append(filtered, q)
		}
	}
	return filtered
}
