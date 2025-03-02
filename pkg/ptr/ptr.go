package ptr

// To returns a pointer to the given value.
func To[T any](v T) *T {
	return &v
}
