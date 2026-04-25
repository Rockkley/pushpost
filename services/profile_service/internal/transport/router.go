package transport

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	handlerhttp "github.com/rockkley/pushpost/services/common_service/http"
	"github.com/rockkley/pushpost/services/common_service/httplog"
	myHTTP "github.com/rockkley/pushpost/services/profile_service/internal/transport/http"
)

func NewRouter(log *slog.Logger, h *myHTTP.ProfileHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(httplog.Logger(log))
	r.Use(chimiddleware.Recoverer)

	r.Route("/profiles", func(r chi.Router) {
		r.Get("/by-username/{username}", handlerhttp.MakeHandler(h.GetByUsername))
		r.Patch("/me", handlerhttp.MakeHandler(h.UpdateMe))
	})

	return r
}
