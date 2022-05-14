package math

func Min[T int | float64](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T int | float64](a, b T) T {
	if a > b {
		return a
	}
	return b
}
