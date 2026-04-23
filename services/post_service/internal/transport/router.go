package transport

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	handlerhttp "github.com/rockkley/pushpost/services/common_service/http"
	"github.com/rockkley/pushpost/services/common_service/httplog"
	commonmiddleware "github.com/rockkley/pushpost/services/common_service/middleware"
	myHTTP "github.com/rockkley/pushpost/services/post_service/internal/transport/http"
)

func NewRouter(log *slog.Logger, h *myHTTP.PostHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(httplog.Logger(log))
	r.Use(chimiddleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Group(func(r chi.Router) {
		r.Use(commonmiddleware.RequireUserID)

		r.Route("/posts", func(r chi.Router) {
			r.Post("/", handlerhttp.MakeHandler(h.CreatePost))
			r.Get("/feed", handlerhttp.MakeHandler(h.GetFeed))
			r.Get("/by-user/{userID}", handlerhttp.MakeHandler(h.GetUserPosts))
			r.Get("/{postID}", handlerhttp.MakeHandler(h.GetPostByID))
			r.Delete("/{postID}", handlerhttp.MakeHandler(h.DeletePost))
		})
	})

	return r
}
