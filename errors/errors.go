package errors

import (
	"fmt"
	"net/http"
)

type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func New(code int, message string) error {
	return &HTTPError{
		Code:    code,
		Message: message,
	}
}

func (e *HTTPError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if text := http.StatusText(e.Code); text != "" {
		return text
	}
	return fmt.Sprintf("http error with code: %d", e.Code)
}

func (e *HTTPError) StatusCode() int {
	return e.Code
}
