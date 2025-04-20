package position

import (
	"math"
	"strings"

	"github.com/metatube-community/metatube-sdk-go/common/cluster"
)

var (
	_ cluster.Locatable[Vector, float64]                          = (*Vector)(nil)
	_ cluster.WeightedLocatable[WeightedVector, float64, float64] = (*WeightedVector)(nil)
)

type Vector struct {
	vector []Position
}

func NewVector(pos ...Position) Vector {
	return Vector{append([]Position(nil), pos...)}
}

func (v Vector) At(i int) Position {
	if i < 0 || i >= len(v.vector) {
		panic("index out of range")
	}
	return v.vector[i]
}

func (v Vector) Dim() int {
	return len(v.vector)
}

func (v Vector) DistanceTo(o Vector) float64 {
	if v.Dim() != o.Dim() {
		panic("dimension mismatch")
	}
	switch v.Dim() {
	case 0:
		return 0.0
	case 1:
		return v.vector[0].DistanceTo(o.vector[0])
	case 2:
		return math.Hypot(
			v.vector[0].DistanceTo(o.vector[0]),
			v.vector[1].DistanceTo(o.vector[1]),
		)
	default:
		var sum float64
		for i := range v.vector {
			d := v.vector[i].DistanceTo(o.vector[i])
			sum += d * d
		}
		return math.Sqrt(sum)
	}
}

func (v Vector) IsValid() bool {
	for _, p := range v.vector {
		if !p.IsValid() {
			return false
		}
	}
	return true
}

func (v Vector) Select(dims ...int) Vector {
	vec := make([]Position, len(dims))
	for i, d := range dims {
		if d < 0 || d >= len(v.vector) {
			panic("index out of range")
		}
		vec[i] = v.vector[d]
	}
	return Vector{vec}
}

func (v Vector) String() string {
	var parts []string
	for _, p := range v.vector {
		parts = append(parts, p.String())
	}
	return "(" + strings.Join(parts, ",") + ")"
}

type WeightedVector struct {
	Vector
	weight float64
}

func NewWeightedVector(vector Vector, weight float64) WeightedVector {
	if weight < 0 {
		panic("weight must be non-negative")
	}
	return WeightedVector{
		Vector: vector,
		weight: weight,
	}
}

func (v WeightedVector) DistanceTo(o WeightedVector) float64 {
	return v.Vector.DistanceTo(o.Vector)
}

func (v WeightedVector) Weight() float64 {
	return v.weight
}
