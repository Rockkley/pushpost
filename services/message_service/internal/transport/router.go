package transport

import (
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	handlerhttp "github.com/rockkley/pushpost/services/common_service/http"
	"github.com/rockkley/pushpost/services/common_service/httplog"
	commonmiddleware "github.com/rockkley/pushpost/services/common_service/middleware"
	myHTTP "github.com/rockkley/pushpost/services/message_service/internal/transport/http"
	"log/slog"
)

func NewRouter(log *slog.Logger, h *myHTTP.MessageHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(httplog.Logger(log))
	r.Use(chimiddleware.Recoverer)
	r.Use(commonmiddleware.RequireUserID)

	r.Route("/messages", func(r chi.Router) {
		r.Post("/", handlerhttp.MakeHandler(h.SendMessage))
		r.Get("/unread/count", handlerhttp.MakeHandler(h.GetUnreadCount))
		r.Get("/unread", handlerhttp.MakeHandler(h.GetUnreadMessages))
		r.Get("/{userID}", handlerhttp.MakeHandler(h.GetConversation))
		r.Patch("/{userID}/read-all", handlerhttp.MakeHandler(h.MarkAllAsRead))
		r.Patch("/{messageID}/read", handlerhttp.MakeHandler(h.MarkAsRead))
	})

	return r
}
