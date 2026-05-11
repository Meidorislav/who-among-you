package httpapi

import "github.com/go-chi/chi/v5"

// NewRouter wires all HTTP and ws routes onto a new chi router.
func NewRouter(h *Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/health", h.Health)

	r.Post("/api/lobby", h.CreateLobby)
	r.Post("/api/lobby/join", h.JoinLobby)

	r.Get("/ws", h.WS)

	return r
}
