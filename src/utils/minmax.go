package utils

// Min calculates the minimum between two unsigned integers (golang has no such function)
func Min[T Value | int | float64 | uint32 | int64](x, y T) T {
	if x < y {
		return x
	}
	return y
}

// Max calculates the max between two unsigned integers (golang has no such function)
func Max[T Value | int | float64 | uint32 | int64](x, y T) T {
	if x > y {
		return x
	}
	return y
}
