package httperror

import (
	"encoding/json"
	"errors"
	"github.com/rockkley/pushpost/internal/apperror"
	"log/slog"
	"net/http"
)

type ErrorResponse struct {
	Code   string            `json:"code"`
	Field  string            `json:"field,omitempty"`
	Fields map[string]string `json:"fields,omitempty"`
}

func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	var appErr apperror.AppError
	if errors.As(err, &appErr) {
		handleAppError(w, r, appErr)
		return
	}

	slog.Error("unexpected error",
		slog.Any("error", err),
		slog.String("path", r.URL.Path),
		slog.String("method", r.Method),
	)

	_ = WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Code: apperror.CodeInternalError})
}

func handleAppError(w http.ResponseWriter, r *http.Request, appErr apperror.AppError) {
	var fieldErrors interface {
		Fields() map[string]string
	}
	switch appErr.Type() {
	case apperror.ErrorTypeClient, apperror.ErrorTypeValidation:
		slog.Debug("client error",
			slog.String("code", appErr.Code()),
			slog.String("field", appErr.Field()),
			slog.String("path", r.URL.Path),
		)
		response := ErrorResponse{
			Code:  appErr.Code(),
			Field: appErr.Field(),
		}
		if errors.As(appErr, &fieldErrors) {
			response.Fields = fieldErrors.Fields()
		}
		_ = WriteJSON(w, appErr.HTTPStatus(), response)
	case apperror.ErrorTypeServer:
		slog.Error("server error",
			slog.Any("error", appErr),
			slog.Any("cause", appErr.Unwrap()),
			slog.String("code", appErr.Code()),
			slog.String("path", r.URL.Path),
			slog.String("method", r.Method),
		)
		_ = WriteJSON(w, appErr.HTTPStatus(), ErrorResponse{Code: appErr.Code()})
	}
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
