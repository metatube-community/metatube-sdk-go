package cluster

import (
	"cmp"
)

// Locatable is a generic interface for any type that can
// measure distance to another value of the same type.
type Locatable[T any, R cmp.Ordered] interface {
	DistanceTo(T) R
}

// Weighted represents a value that carries a weight.
type Weighted[W cmp.Ordered] interface {
	Weight() W
}

// WeightedLocatable combines both Locatable and Weighted behaviors.
type WeightedLocatable[T any, R cmp.Ordered, W cmp.Ordered] interface {
	Locatable[T, R]
	Weighted[W]
}
