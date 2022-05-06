package provider

import (
	"errors"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrNotImplemented = errors.New("not implemented")
	ErrNotSupported   = errors.New("not supported")
)
