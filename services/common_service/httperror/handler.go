package httperror

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/ctxlog"
)

type ErrorResponse struct {
	Code   string            `json:"code"`
	Field  string            `json:"field,omitempty"`
	Fields map[string]string `json:"fields,omitempty"`
}

// HandleError maps an error to an HTTP response. If the error is an AppError, it uses its HTTP status and code;
// otherwise, it returns a generic 500 error.
func HandleError(w http.ResponseWriter, r *http.Request, err error) error {
	log := ctxlog.From(r.Context())

	var appErr apperror.AppError
	if !errors.As(err, &appErr) {
		log.Error("unhandled error", slog.Any("error", err))
		if writeErr := WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Code: apperror.CodeInternalError}); writeErr != nil {
			log.Error("failed to write error response",
				slog.Int("status", http.StatusInternalServerError),
				slog.String("code", apperror.CodeInternalError),
				slog.Any("error", writeErr),
			)
			return fmt.Errorf("write internal error response: %w", writeErr)
		}
		return nil
	}

	status := appErr.HTTPStatus()

	if status >= 500 {
		log.Error("server error",
			slog.Int("status", status),
			slog.String("code", appErr.Code()),
			slog.Any("cause", appErr.Unwrap()),
		)
	} else {
		log.Debug("client error",
			slog.Int("status", status),
			slog.String("code", appErr.Code()),
			slog.String("field", appErr.Field()),
		)
	}

	resp := ErrorResponse{
		Code:   appErr.Code(),
		Field:  appErr.Field(),
		Fields: appErr.Fields(),
	}
	if err = WriteJSON(w, status, resp); err != nil {
		log.Error("failed to write error response",
			slog.Int("status", status),
			slog.String("code", appErr.Code()),
			slog.Any("error", err),
		)
		return fmt.Errorf("write error response: %w", err)
	}

	return nil
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
