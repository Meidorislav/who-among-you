package main

import (
	"net/http"
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

	http.ListenAndServe(":8080", r)
}
