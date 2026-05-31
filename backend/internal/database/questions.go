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
	rows, err := db.Query(ctx, "SELECT text_en, text_ru FROM questions")
	if err != nil {
		return nil, fmt.Errorf("query questions: %w", err)
	}
	defer rows.Close()

	var questions []game.Question
	for rows.Next() {
		var q game.Question
		if err := rows.Scan(&q.TextEn, &q.TextRu); err != nil {
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

func (p *PostgresQuestions) Len() int {
	return len(p.pool)
}

func (p *PostgresQuestions) Next() game.Question {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.pool[p.rng.Intn(len(p.pool))]
}
