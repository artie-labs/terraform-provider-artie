package lib

func RemovePtr[T any](ptr *T) T {
	var zero T
	if ptr == nil {
		return zero
	}

	return *ptr
}
