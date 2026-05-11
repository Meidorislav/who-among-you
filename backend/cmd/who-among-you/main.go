package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
	"who-among-you/internal/httpapi"
	"who-among-you/internal/lobby"

	"github.com/go-chi/chi/v5"
)

func main() {
	lobbies := lobby.InitLobbies()
	handler := httpapi.NewHandler(lobbies)

	r := chi.NewRouter()

	r.Get("/health", handler.Health)
	r.Post("/api/lobby", handler.CreateLobby)
	r.Post("/api/lobby/join", handler.JoinLobby)

	srv := &http.Server{Addr: ":8080", Handler: r}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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
