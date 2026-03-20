package transport

import (
	handlerhttp "github.com/rockkley/pushpost/services/common_service/http"
	"github.com/rockkley/pushpost/services/common_service/httplog"
	"log/slog"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	myHTTP "github.com/rockkley/pushpost/services/user_service/internal/transport/http"
)

func NewRouter(log *slog.Logger, userHandler *myHTTP.UserHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(httplog.Logger(log))
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.URLFormat)

	r.Route("/users", func(r chi.Router) {
		r.Post("/", handlerhttp.MakeHandler(userHandler.CreateUser))
		r.Get("/{id}", handlerhttp.MakeHandler(userHandler.GetUserByID))
		r.Get("/by-email", handlerhttp.MakeHandler(userHandler.GetUserByEmail))
		r.Get("/by-username/{username}", handlerhttp.MakeHandler(userHandler.GetUserByUsername))
	})

	return r
}
