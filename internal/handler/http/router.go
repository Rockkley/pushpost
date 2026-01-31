package http

import "github.com/go-chi/chi/v5"

func NewRouter(authMW *AuthMiddleware, authHandler *AuthHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", MakeHandler(authHandler.Register))
		r.Post("/login", MakeHandler(authHandler.Login))
	})

	r.Route("/api/user", func(r chi.Router) {
		r.Use(authMW.RequireAuth)
	})
	return r
}
