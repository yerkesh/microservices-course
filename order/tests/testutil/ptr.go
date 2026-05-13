package testutil

// Ptr возвращает указатель на переданное значение.
func Ptr[T any](value T) *T {
	return &value
}
