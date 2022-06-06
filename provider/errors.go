package provider

import (
	"net/http"

	"github.com/javtube/javtube-sdk-go/errors"
)

var (
	ErrInvalidID          = &errors.HTTPError{Code: http.StatusBadRequest, Message: "invalid id"}
	ErrInvalidURL         = &errors.HTTPError{Code: http.StatusBadRequest, Message: "invalid url"}
	ErrInvalidKeyword     = &errors.HTTPError{Code: http.StatusBadRequest, Message: "invalid keyword"}
	ErrInfoNotFound       = &errors.HTTPError{Code: http.StatusNotFound, Message: "info not found"}
	ErrProviderNotFound   = &errors.HTTPError{Code: http.StatusNotFound, Message: "provider not found"}
	ErrIncompleteMetadata = &errors.HTTPError{Code: http.StatusInternalServerError, Message: "incomplete metadata"}
)
