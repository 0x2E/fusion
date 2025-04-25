package ptr

// To returns a pointer to the given value.
func To[T any](v T) *T {
	return &v
}

// From returns the value pointed to by the given pointer.
// If the pointer is nil, it returns the zero value of the type.
func From[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}
	return *v
}
