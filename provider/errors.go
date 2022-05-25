package provider

import (
	"errors"
)

var (
	ErrInvalidID       = errors.New("invalid id")
	ErrInvalidURL      = errors.New("invalid url")
	ErrInvalidKeyword  = errors.New("invalid keyword")
	ErrInvalidMetadata = errors.New("invalid metadata")
	ErrNotFound        = errors.New("not found")
	ErrNotSupported    = errors.New("not supported")
)
