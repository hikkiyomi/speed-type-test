package internal

func Filter[T any](elem []T, f func(e T) bool) []T {
	result := make([]T, 0)

	for _, e := range elem {
		if f(e) {
			result = append(result, e)
		}
	}

	return result
}
