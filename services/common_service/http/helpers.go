package http

import (
	"github.com/rockkley/pushpost/services/common_service/ctxlog"
	"github.com/rockkley/pushpost/services/common_service/httperror"
	"log/slog"
	"net/http"
)

type APIFunc func(w http.ResponseWriter, r *http.Request) error

func MakeHandler(h APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			if handleErr := httperror.HandleError(w, r, err); handleErr != nil {
				ctxlog.From(r.Context()).Error("failed to handle api error", slog.Any("error", handleErr))
			}

		}

	}
}
