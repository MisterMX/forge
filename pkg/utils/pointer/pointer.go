package pointer

// Return the value of ptr or fallback if ptr is nil.
func Deref[T any](ptr *T, fallback T) T {
	if ptr == nil {
		return fallback
	}
	return *ptr
}
