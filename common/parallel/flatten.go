package parallel

func Flatten[T any](input [][]T) []T {
	var result []T
	for _, inner := range input {
		result = append(result, inner...)
	}
	return result
}
