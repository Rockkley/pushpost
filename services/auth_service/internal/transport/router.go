package transport

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	handlerhttp "github.com/rockkley/pushpost/internal/handler/http"
	"github.com/rockkley/pushpost/pkg/httplog"
	myHTTP "github.com/rockkley/pushpost/services/auth_service/internal/transport/http"
	authmiddleware "github.com/rockkley/pushpost/services/auth_service/internal/transport/http/middleware"
)

func NewRouter(
	log *slog.Logger,
	authMW *authmiddleware.AuthMiddleware,
	authHandler *myHTTP.AuthHandler,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{ // FIXME
		AllowedOrigins:   []string{"http://localhost:63342"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Use(chimiddleware.RequestID)
	r.Use(httplog.Logger(log))
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.URLFormat)

	r.Post("/register", handlerhttp.MakeHandler(authHandler.Register))
	r.Post("/login", handlerhttp.MakeHandler(authHandler.Login))

	r.Route("/ra", func(r chi.Router) {
		r.Use(authMW.RequireAuth)
		r.Post("/logout", handlerhttp.MakeHandler(authHandler.Logout))
	})

	return r
}
