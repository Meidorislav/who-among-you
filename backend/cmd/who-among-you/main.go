package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"who-among-you/internal/database"
	"who-among-you/internal/game"
	"who-among-you/internal/httpapi"
	"who-among-you/internal/lobby"
	"who-among-you/internal/ws"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var questions game.QuestionSource
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		db, err := database.Connect(ctx, dsn)
		if err != nil {
			log.Fatalf("database: %v", err)
		}
		defer db.Close()

		pq, err := database.NewPostgresQuestions(ctx, db)
		if err != nil {
			log.Fatalf("load questions: %v", err)
		}
		log.Printf("loaded %d questions from database", pq.Len(nil))
		questions = pq
	} else {
		log.Println("DATABASE_URL not set, using mock questions")
		questions = game.NewMockQuestions()
	}

	lobbies := lobby.InitLobbies()
	hub := ws.NewHub()
	games := game.NewManager(questions, hub)
	handler := httpapi.NewHandler(lobbies, hub, games)
	hub.SetHandler(handler)
	go hub.Run()

	srv := &http.Server{Addr: ":8080", Handler: httpapi.NewRouter(handler)}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()
	log.Println("server listening on :8080")

	<-ctx.Done()
	log.Println("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
	log.Println("server stopped")
}
