package position

var (
	_ = AverageVector
	_ = WeightedAverageVector
)

func AverageVector(vectors []Vector) Vector {
	return WeightedAverageBy(vectors,
		func(v Vector) WeightedVector {
			return WeightedVector{v, 1.0}
		},
	)
}

func WeightedAverageVector(vectors []WeightedVector) Vector {
	return WeightedAverageBy(vectors,
		func(v WeightedVector) WeightedVector {
			return v
		},
	)
}

func WeightedAverageBy[T any](items []T, get func(T) WeightedVector) Vector {
	if len(items) == 0 {
		return Vector{}
	}

	dim := get(items[0]).Dim()
	sum := make([]Position, dim)
	var totalWeight float64

	for _, item := range items {
		wv := get(item)

		if wv.Dim() != dim {
			panic("dimension mismatch")
		}

		for i := 0; i < dim; i++ {
			sum[i] += wv.At(i) * Position(wv.Weight())
		}
		totalWeight += wv.Weight()
	}

	if totalWeight <= 0 {
		return Vector{}
	}

	for i := 0; i < dim; i++ {
		sum[i] /= Position(totalWeight)
	}

	return Vector{sum}
}
