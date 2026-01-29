package http

import "github.com/go-chi/chi/v5"

func NewRouter(h *AuthHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", MakeHandler(h.Register))
		//r.Post("/login", Make(h.Login))
	})

	return r
}
