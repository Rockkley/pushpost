package transport

import (
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rockkley/pushpost/services/api_gateway/internal/middleware"
	"github.com/rockkley/pushpost/services/common_service/httplog"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"strings"
)

type Proxies struct {
	Auth           *httputil.ReverseProxy
	User           *httputil.ReverseProxy
	Friendship     *httputil.ReverseProxy
	UserByUsername *httputil.ReverseProxy
}

func RewriteUsernameToPath(path string) string {
	username := strings.TrimPrefix(path, "/")
	return "/users/by-username/" + username
}

func NewRouter(
	log *slog.Logger,
	authMW *middleware.AuthMiddleware, p Proxies) *chi.Mux {

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"}, // FIXME: lock down to real domains
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(chimiddleware.RequestID)
	r.Use(httplog.Logger(log))
	r.Use(chimiddleware.Recoverer)

	r.Handle("/auth/*", http.HandlerFunc(p.Auth.ServeHTTP))

	r.Group(func(r chi.Router) {
		r.Use(authMW.RequireAuth)

		r.Handle("/users/*", http.HandlerFunc(p.User.ServeHTTP))
		r.Handle("/friends", http.HandlerFunc(p.Friendship.ServeHTTP))
		r.Handle("/friends/*", http.HandlerFunc(p.Friendship.ServeHTTP))
	})

	r.Get("/{username}/", p.UserByUsername.ServeHTTP)

	return r
}
