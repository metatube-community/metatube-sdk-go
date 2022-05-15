package provider

import (
	"errors"
)

var (
	ErrInvalidID       = errors.New("invalid id")
	ErrInvalidKeyword  = errors.New("invalid keyword")
	ErrInvalidMetadata = errors.New("invalid metadata")
	ErrNotFound        = errors.New("not found")
	ErrNotSupported    = errors.New("not supported")
)
