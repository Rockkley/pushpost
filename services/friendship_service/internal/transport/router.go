package router

import (
	"github.com/rockkley/pushpost/services/friendship_service/internal/transport/http/middleware"
	"log/slog"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	handlerhttp "github.com/rockkley/pushpost/services/common_service/http"
	"github.com/rockkley/pushpost/services/common_service/httplog"
	myHTTP "github.com/rockkley/pushpost/services/friendship_service/internal/transport/http"
)

func NewRouter(log *slog.Logger, h *myHTTP.FriendshipHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(httplog.Logger(log))
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.RequireUserID)

	r.Post("/friends/requests", handlerhttp.MakeHandler(h.SendRequest))
	r.Get("/friends/requests/incoming", handlerhttp.MakeHandler(h.GetIncomingRequests))
	r.Post("/friends/requests/{senderID}/accept", handlerhttp.MakeHandler(h.AcceptRequest))
	r.Post("/friends/requests/{senderID}/reject", handlerhttp.MakeHandler(h.RejectRequest))
	r.Delete("/friends/requests/{receiverID}", handlerhttp.MakeHandler(h.CancelRequest))
	r.Delete("/friends/{userID}", handlerhttp.MakeHandler(h.DeleteFriendship))
	r.Get("/friends", handlerhttp.MakeHandler(h.GetFriendIDs))
	r.Get("/friends/{userID}/status", handlerhttp.MakeHandler(h.AreFriends))
	r.Get("/friends/{userID}/relationship", handlerhttp.MakeHandler(h.GetRelationship))

	return r
}
