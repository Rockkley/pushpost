package transport

import (
	handlerhttp "github.com/rockkley/pushpost/services/common_service/http"
	"github.com/rockkley/pushpost/services/common_service/httplog"
	"log/slog"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	myHTTP "github.com/rockkley/pushpost/services/auth_service/internal/transport/http"
	authmiddleware "github.com/rockkley/pushpost/services/auth_service/internal/transport/http/middleware"
	"github.com/rockkley/pushpost/services/common_service/metrics"
)

func NewRouter(
	log *slog.Logger,
	authMW *authmiddleware.AuthMiddleware,
	authHandler *myHTTP.AuthHandler,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(httplog.Logger(log))
	r.Use(chimiddleware.Recoverer)
	r.Use(metrics.Middleware("auth-service"))
	r.Use(chimiddleware.URLFormat)
	r.Handle("/metrics", metrics.Handler())

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", handlerhttp.MakeHandler(authHandler.Register))
		r.Post("/login", handlerhttp.MakeHandler(authHandler.Login))
		r.Post("/verify-email", handlerhttp.MakeHandler(authHandler.VerifyEmail))
		r.Post("/resend-otp", handlerhttp.MakeHandler(authHandler.ResendOTP))

		r.Group(func(r chi.Router) {
			r.Use(authMW.RequireAuth)
			r.Post("/logout", handlerhttp.MakeHandler(authHandler.Logout))
		})
	})

	return r
}
