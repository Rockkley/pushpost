package http

//var errorStatusMap = map[error]int{
//	entity.ErrUserNotFound:      http.StatusNotFound,
//	entity.ErrMessageNotFound:   http.StatusNotFound,
//	entity.ErrCannotMessageSelf: http.StatusBadRequest,
//	entity.ErrMessageTooLong:    http.StatusBadRequest,
//	entity.ErrMessageEmpty:      http.StatusBadRequest,
//	entity.ErrReceiverDeleted:   http.StatusBadRequest,
//	entity.ErrSenderDeleted:     http.StatusForbidden,
//}

type HTTPError struct {
	Status int
	Code   string
	Field  string
}

func (e *HTTPError) Error() string {
	return e.Code
}
