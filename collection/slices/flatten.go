package slices

// Flatten flattens a 2D slice into a 1D
// slice by merging all inner slices.
func Flatten[T any](input [][]T) []T {
	var result []T
	for _, inner := range input {
		result = append(result, inner...)
	}
	return result
}
