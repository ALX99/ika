package config

func firstSet[T any](vals ...Nullable[T]) Nullable[T] {
	for _, v := range vals {
		if v.set {
			return v
		}
	}
	return Nullable[T]{}
}
