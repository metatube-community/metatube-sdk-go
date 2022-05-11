package provider

import (
	"errors"
)

var (
	ErrInvalidID      = errors.New("invalid id")
	ErrNotFound       = errors.New("not found")
	ErrNotImplemented = errors.New("not implemented")
	ErrNotSupported   = errors.New("not supported")
)
