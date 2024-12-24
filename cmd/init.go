package cmd

import (
	"go.uber.org/automaxprocs/maxprocs"
)

func init() {
	maxprocs.Set(maxprocs.Logger(func(string, ...any) {}))
}
