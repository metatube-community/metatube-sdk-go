package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// HTTPError implements error interface with HTTP status code.
type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *HTTPError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if text := http.StatusText(e.Code); text != "" {
		return text
	}
	return fmt.Sprintf("error code: %d", e.Code)
}

func (e *HTTPError) StatusCode() int {
	return e.Code
}

func (e *HTTPError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"code":    e.Code,
		"message": e.Error(),
	})
}

func New(code int, message string) error {
	return &HTTPError{
		Code:    code,
		Message: message,
	}
}

func FromCode(code int) error {
	return &HTTPError{
		Code:    code,
		Message: http.StatusText(code),
	}
}

var _ error = (*HTTPError)(nil)
