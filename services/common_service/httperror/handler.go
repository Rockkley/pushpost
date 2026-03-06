package httperror

import (
	"encoding/json"
	"errors"
	"github.com/rockkley/pushpost/services/common_service/ctxlog"
	"log/slog"
	"net/http"

	"github.com/rockkley/pushpost/services/common_service/apperror"
)

type ErrorResponse struct {
	Code   string            `json:"code"`
	Field  string            `json:"field,omitempty"`
	Fields map[string]string `json:"fields,omitempty"`
}

func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	log := ctxlog.From(r.Context())

	var appErr apperror.AppError
	if errors.As(err, &appErr) {
		handleAppError(w, log, appErr)
		return
	}

	log.Error("unhandled error", slog.Any("error", err))
	WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Code: apperror.CodeInternalError})
}

func handleAppError(w http.ResponseWriter, log *slog.Logger, appErr apperror.AppError) {
	switch appErr.Type() {

	case apperror.ErrorTypeClient, apperror.ErrorTypeValidation:
		log.Debug("client error",
			slog.String("code", appErr.Code()),
			slog.String("field", appErr.Field()),
			slog.Int("status", appErr.HTTPStatus()),
		)
		resp := ErrorResponse{Code: appErr.Code(), Field: appErr.Field()}
		if appErr.Fields() != nil {
			resp.Fields = appErr.Fields()
		}
		WriteJSON(w, appErr.HTTPStatus(), resp)

	case apperror.ErrorTypeServer:
		log.Error("server error",
			slog.String("code", appErr.Code()),
			slog.Any("cause", appErr.Unwrap()),
		)
		WriteJSON(w, appErr.HTTPStatus(), ErrorResponse{Code: appErr.Code()})

	default:
		log.Error("unhandled apperror type",
			slog.Int("type", int(appErr.Type())),
			slog.String("code", appErr.Code()),
			slog.Any("error", appErr),
		)
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Code: apperror.CodeInternalError})
	}
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
