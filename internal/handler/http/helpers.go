package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"github.com/rockkley/pushpost/internal/domain"
	"net/http"
)

type APIError struct {
	StatusCode int `json:"statusCode"`
	Msg        any `json:"msg"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("api error: %d", e.StatusCode)
}

func NewApiError(statusCode int, err error) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Msg:        err.Error()}
}

func InvalidRequestData(errors map[string]string) APIError {
	return APIError{
		StatusCode: http.StatusUnprocessableEntity,
		Msg:        errors,
	}
}

func InvalidJSON() APIError {
	return APIError{
		StatusCode: http.StatusBadRequest,
		Msg:        fmt.Errorf("invalid json request data"),
	}
}

type APIFunc func(w http.ResponseWriter, r *http.Request) error

func MakeHandler(h APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			var apiErr APIError
			if errors.As(err, &apiErr) {
				WriteJSON(w, apiErr.StatusCode, apiErr)
				return
			}
			var de domain.DomainError
			if errors.As(err, &de) {
				WriteJSON(w, de.HTTPStatus(), ErrorResponse{
					Field: de.GetField(),
					Code:  de.Code(),
				})
			} else {
				WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Code: "internal_error"})
			}
			log.Errorf("HTTP API error: %v, path: %s", err, r.URL.Path)

		}
	}
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)

}
