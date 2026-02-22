package transport

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rockkley/pushpost/internal/handler/httperror"
	myHTTP "github.com/rockkley/pushpost/services/user_service/internal/transport/http"
	"github.com/rockkley/pushpost/services/user_service/internal/transport/http/middleware"
)

func NewRouter(log *slog.Logger, userHandler *myHTTP.UserHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(middleware.Logger(log))
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.URLFormat)

	r.Post("/user", MakeHandler(userHandler.CreateUser))
	r.Post("/users/authenticate-user", MakeHandler(userHandler.AuthenticateUser))

	return r
}

type APIFunc func(w http.ResponseWriter, r *http.Request) error

func MakeHandler(h APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			httperror.HandleError(w, r, err)
		}
	}
}
