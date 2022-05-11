package provider

import (
	"errors"
)

var (
	ErrInvalidID      = errors.New("invalid id")
	ErrInvalidKeyword = errors.New("invalid keyword")
	ErrNotFound       = errors.New("not found")
	ErrNotSupported   = errors.New("not supported")
)
