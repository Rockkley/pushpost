package transport

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rockkley/pushpost/internal/handler/httperror"
	myHTTP "github.com/rockkley/pushpost/services/user_service/internal/transport/http"
	"net/http"
)

type APIFunc func(w http.ResponseWriter, r *http.Request) error

func NewRouter(userHandler *myHTTP.UserHandler) *chi.Mux {
	r := chi.NewRouter()
	//log := logger.SetupLogger("local") // fixme

	r.Use(middleware.RequestID)
	//r.Use(myMW.New(log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)

	r.Post("/user", MakeHandler(userHandler.CreateUser))

	return r
}

func MakeHandler(h APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			httperror.HandleError(w, r, err)
		}

	}
}
