package router

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	handlerhttp "github.com/rockkley/pushpost/services/common_service/http"
	"github.com/rockkley/pushpost/services/common_service/httplog"
	authmw "github.com/rockkley/pushpost/services/common_service/middleware"
	myHTTP "github.com/rockkley/pushpost/services/friendship_service/internal/transport/http"
)

func NewRouter(
	log *slog.Logger,
	authMW *authmw.JwtAuthMiddleware,
	h *myHTTP.FriendshipHandler,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{ // FIXME: lock down in production
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(chimiddleware.RequestID)
	r.Use(httplog.Logger(log))
	r.Use(chimiddleware.Recoverer)

	r.Group(func(r chi.Router) {
		r.Use(authMW.RequireAuth)

		r.Post("/friends/requests", handlerhttp.MakeHandler(h.SendRequest))
		r.Post("/friends/requests/{senderID}/accept", handlerhttp.MakeHandler(h.AcceptRequest))
		r.Post("/friends/requests/{senderID}/reject", handlerhttp.MakeHandler(h.RejectRequest))
		r.Delete("/friends/requests/{receiverID}", handlerhttp.MakeHandler(h.CancelRequest))

		// Established friendships
		r.Delete("/friends/{friendID}", handlerhttp.MakeHandler(h.DeleteFriendship))
		r.Get("/friends", handlerhttp.MakeHandler(h.GetFriendIDs))
		r.Get("/friends/{friendID}/status", handlerhttp.MakeHandler(h.AreFriends))
	})

	return r
}
