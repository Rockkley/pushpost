package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	myMW "github.com/rockkley/pushpost/internal/handler/http/middleware"
	http2 "github.com/rockkley/pushpost/internal/services/auth_service/transport/http"
	"github.com/rockkley/pushpost/pkg/logger"
)

func NewRouter(authMW *myMW.AuthMiddleware, authHandler *http2.AuthHandler) *chi.Mux {
	r := chi.NewRouter()
	log := logger.SetupLogger("local") // fixme

	r.Use(middleware.RequestID)
	r.Use(myMW.New(log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)

	r.Route("/api/user_service-service", func(r chi.Router) {
		r.Post("/register", MakeHandler(authHandler.Register))
		r.Post("/login", MakeHandler(authHandler.Login))
	})

	r.Route("/api", func(r chi.Router) {
	}
	r.Route("/api/user_service", func(r chi.Router) {
		r.Use(authMW.RequireAuth)
		r.Post("/logout", MakeHandler(authHandler.Logout))
	})

	r.Route()
	return r
}
