package secureerr

import (
	"errors"
	"fmt"
	"net/http"
)


type Error struct {
	Code     string            `json:"code"`
	UserMsg  string            `json:"error"`
	Status   int               `json:"-"`
	Internal error             `json:"-"`
	Metadata map[string]string `json:"-"`
}

func New(status int, code, userMsg string, internal error, metadata map[string]string) *Error {
	if status == 0 {
		status = http.StatusInternalServerError
	}
	if userMsg == "" {
		userMsg = http.StatusText(status)
	}
	if code == "" {
		code = "INTERNAL_ERROR"
	}

	return &Error{
		Code:     code,
		UserMsg:  userMsg,
		Status:   status,
		Internal: internal,
		Metadata: metadata,
	}
}

func BadRequest(code, userMsg string, internal error, metadata map[string]string) *Error {
	return New(http.StatusBadRequest, code, userMsg, internal, metadata)
}

func NotFound(code, userMsg string, internal error, metadata map[string]string) *Error {
	return New(http.StatusNotFound, code, userMsg, internal, metadata)
}

func Conflict(code, userMsg string, internal error, metadata map[string]string) *Error {
	return New(http.StatusConflict, code, userMsg, internal, metadata)
}

func Internal(code, userMsg string, internal error, metadata map[string]string) *Error {
	return New(http.StatusInternalServerError, code, userMsg, internal, metadata)
}

func (e *Error) Error() string {
	if e == nil {
		return http.StatusText(http.StatusInternalServerError)
	}
	return e.UserMsg
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Internal
}

func (e *Error) StatusCode() int {
	if e == nil || e.Status == 0 {
		return http.StatusInternalServerError
	}
	return e.Status
}

func (e *Error) LogString() string {
	if e == nil {
		return "status=500 code=INTERNAL_ERROR msg=\"Internal Server Error\""
	}

	return fmt.Sprintf(
		"status=%d code=%s msg=%q cause=%v metadata=%v",
		e.StatusCode(),
		e.Code,
		e.UserMsg,
		e.Internal,
		e.Metadata,
	)
}

func FromError(err error) *Error {
	if err == nil {
		return nil
	}

	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}

	return Internal("INTERNAL_ERROR", http.StatusText(http.StatusInternalServerError), err, nil)
}
