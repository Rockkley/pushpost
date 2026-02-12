package http

import (
	entity2 "github.com/rockkley/pushpost/internal/services/message/entity"
	"github.com/rockkley/pushpost/internal/services/user_service/internal/entity"
	"net/http"
)

var errorStatusMap = map[error]int{
	entity.ErrUserNotFound:       http.StatusNotFound,
	entity2.ErrMessageNotFound:   http.StatusNotFound,
	entity2.ErrCannotMessageSelf: http.StatusBadRequest,
	entity2.ErrMessageTooLong:    http.StatusBadRequest,
	entity2.ErrMessageEmpty:      http.StatusBadRequest,
	entity2.ErrReceiverDeleted:   http.StatusBadRequest,
	entity2.ErrSenderDeleted:     http.StatusForbidden,
}

type HTTPError struct {
	Status int
	Code   string
	Field  string
}

func (e *HTTPError) Error() string {
	return e.Code
}
