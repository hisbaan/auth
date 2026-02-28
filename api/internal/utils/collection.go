package utils

func Filter[T any](list []T, cond func(T) bool) []T {
	acc := make([]T, 0)
	for _, elem := range list {
		if cond(elem) {
			acc = append(acc, elem)
		}
	}
	return acc
}

func Map[I any, O any](list []I, f func(I) O) []O {
	acc := make([]O, 0)
	for _, elem := range list {
		acc = append(acc, f(elem))
	}
	return acc
}
