package transport

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rockkley/pushpost/internal/handler/httperror"
	"github.com/rockkley/pushpost/pkg/logger"
	myHTTP "github.com/rockkley/pushpost/services/auth_service/internal/transport/http"
	middleware2 "github.com/rockkley/pushpost/services/auth_service/internal/transport/http/middleware"
	"net/http"
)

func NewRouter(authMW *middleware2.AuthMiddleware, authHandler *myHTTP.AuthHandler) *chi.Mux {
	r := chi.NewRouter()
	log := logger.SetupLogger("local") // fixme

	r.Use(middleware.RequestID)
	r.Use(middleware2.New(log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)

	r.Post("/register", MakeHandler(authHandler.Register))
	r.Post("/login", MakeHandler(authHandler.Login))

	r.Route("/api/user_service", func(r chi.Router) {
		r.Use(authMW.RequireAuth)
		r.Post("/logout", MakeHandler(authHandler.Logout))
	})

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
