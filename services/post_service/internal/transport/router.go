package transport

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	handlerhttp "github.com/rockkley/pushpost/services/common_service/http"
	"github.com/rockkley/pushpost/services/common_service/httplog"
	"github.com/rockkley/pushpost/services/common_service/metrics"
	commonmiddleware "github.com/rockkley/pushpost/services/common_service/middleware"
	myHTTP "github.com/rockkley/pushpost/services/post_service/internal/transport/http"
)

func NewRouter(log *slog.Logger, h *myHTTP.PostHandler, sseHandler *myHTTP.FeedSSEHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(httplog.Logger(log))
	r.Use(metrics.Middleware("post-service"))
	r.Use(chimiddleware.Recoverer)
	r.Handle("/metrics", metrics.Handler())

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Group(func(r chi.Router) {
		r.Use(commonmiddleware.RequireUserID)

		r.Route("/posts", func(r chi.Router) {
			r.Post("/", handlerhttp.MakeHandler(h.CreatePost))
			r.Get("/", handlerhttp.MakeHandler(h.GetPostsByIDs)) // GET /posts?ids=id1,id2
			r.Get("/feed", handlerhttp.MakeHandler(h.GetFeed))
			r.Get("/feed/subscribe", sseHandler.Subscribe) // SSE — не MakeHandler, управляет ответом сам
			r.Get("/by-user/{userID}", handlerhttp.MakeHandler(h.GetUserPosts))
			r.Get("/{postID}", handlerhttp.MakeHandler(h.GetPostByID))
			r.Patch("/{postID}", handlerhttp.MakeHandler(h.UpdatePost))
			r.Delete("/{postID}", handlerhttp.MakeHandler(h.DeletePost))
			r.Put("/{postID}/like", handlerhttp.MakeHandler(h.LikePost))
			r.Put("/{postID}/dislike", handlerhttp.MakeHandler(h.DislikePost))
			r.Delete("/{postID}/vote", handlerhttp.MakeHandler(h.RemoveVote))

		})
	})

	return r
}
