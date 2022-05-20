package math

type num interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 |
		float32 | float64
}

func Min[T num](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T num](a, b T) T {
	if a > b {
		return a
	}
	return b
}
