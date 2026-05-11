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
	"who-among-you/internal/ws"
)

func main() {
	lobbies := lobby.InitLobbies()
	hub := ws.NewHub()
	handler := httpapi.NewHandler(lobbies, hub)
	hub.SetHandler(handler)
	go hub.Run()

	srv := &http.Server{Addr: ":8080", Handler: httpapi.NewRouter(handler)}

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
