package http

import (
	"github.com/rockkley/pushpost/internal/handler/httperror"
	"net/http"
)

type APIFunc func(w http.ResponseWriter, r *http.Request) error

func MakeHandler(h APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			httperror.HandleError(w, r, err)
		}

	}
}
