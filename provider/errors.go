package provider

import (
	"errors"
)

var (
	ErrNotImplemented = errors.New("not implemented")
	ErrNotSupported   = errors.New("not supported")
)
