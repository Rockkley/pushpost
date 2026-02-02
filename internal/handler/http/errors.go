package http

import (
	"github.com/rockkley/pushpost/internal/domain"
	"net/http"
)

var errorStatusMap = map[error]int{
	domain.ErrUserNotFound:      http.StatusNotFound,
	domain.ErrMessageNotFound:   http.StatusNotFound,
	domain.ErrCannotMessageSelf: http.StatusBadRequest,
	domain.ErrMessageTooLong:    http.StatusBadRequest,
	domain.ErrMessageEmpty:      http.StatusBadRequest,
	domain.ErrReceiverDeleted:   http.StatusBadRequest,
	domain.ErrSenderDeleted:     http.StatusForbidden,
}

type HTTPError struct {
	Status int
	Code   string
	Field  string
}

func (e *HTTPError) Error() string {
	return e.Code
}
