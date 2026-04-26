package transport

import (
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	handlerhttp "github.com/rockkley/pushpost/services/common_service/http"
	"github.com/rockkley/pushpost/services/common_service/httplog"
	commonmiddleware "github.com/rockkley/pushpost/services/common_service/middleware"
	myHTTP "github.com/rockkley/pushpost/services/notification_service/internal/transport/http"
)

func NewRouter(log *slog.Logger, h *myHTTP.NotificationHandler, sse *myHTTP.SSEHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(httplog.Logger(log))
	r.Use(chimiddleware.Recoverer)
	r.Use(commonmiddleware.RequireUserID)

	r.Route("/notifications", func(r chi.Router) {
		r.Get("/stream", sse.Subscribe)
		r.Group(func(r chi.Router) {
			r.Use(chimiddleware.Timeout(10 * time.Second))
			r.Get("/", handlerhttp.MakeHandler(h.List))
			r.Get("/unread/count", handlerhttp.MakeHandler(h.GetUnreadCount))
			r.Patch("/read-all", handlerhttp.MakeHandler(h.MarkAllAsRead))
			r.Patch("/{notificationID}/read", handlerhttp.MakeHandler(h.MarkAsRead))
			r.Route("/preferences", func(r chi.Router) {
				r.Get("/", handlerhttp.MakeHandler(h.GetPreferences))
				r.Put("/", handlerhttp.MakeHandler(h.SetPreference))
			})
			r.Route("/telegram", func(r chi.Router) {
				r.Post("/link", handlerhttp.MakeHandler(h.GenerateTelegramCode))
				r.Delete("/link", handlerhttp.MakeHandler(h.UnbindTelegram))
			})
		})
	})
	return r
}
