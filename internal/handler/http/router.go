package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	myMW "github.com/rockkley/pushpost/internal/handler/http/middleware"
	"github.com/rockkley/pushpost/pkg/logger"
)

func NewRouter(authMW *myMW.AuthMiddleware, authHandler *AuthHandler) *chi.Mux {
	r := chi.NewRouter()
	log := logger.SetupLogger("local") // fixme

	r.Use(middleware.RequestID)
	r.Use(myMW.New(log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)

	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", MakeHandler(authHandler.Register))
		r.Post("/login", MakeHandler(authHandler.Login))
	})

	r.Route("/api/user", func(r chi.Router) {
		r.Use(authMW.RequireAuth)
	})
	return r
}
