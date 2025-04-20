package slices

import (
	"slices"
)

// Flatten flattens a 2D slice into a 1D slice by merging all inner slices.
func Flatten[E any](s [][]E) []E {
	return slices.Collect(func(yield func(E) bool) {
		for _, i := range s {
			for _, j := range i {
				if !yield(j) {
					return
				}
			}
		}
	})
}
