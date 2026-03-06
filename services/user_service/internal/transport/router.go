package transport

import (
	handlerhttp "github.com/rockkley/pushpost/services/common_service/http"
	"github.com/rockkley/pushpost/services/common_service/httplog"
	"log/slog"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	myHTTP "github.com/rockkley/pushpost/services/user_service/internal/transport/http"
)

func NewRouter(log *slog.Logger, userHandler *myHTTP.UserHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{ // FIXME: lock down in production
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

	r.Post("/user", handlerhttp.MakeHandler(userHandler.CreateUser))
	r.Post("/users/authenticate-user", handlerhttp.MakeHandler(userHandler.AuthenticateUser))
	r.Get("/users/by-email", handlerhttp.MakeHandler(userHandler.GetUserByEmail))

	return r
}
