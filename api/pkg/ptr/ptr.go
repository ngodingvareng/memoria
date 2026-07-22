// Package ptr provides small generic helpers for working with pointers
// to plain values — mainly for bridging optional/nullable fields (e.g.
// entity structs with *string/*int32 fields) without writing one-off
// "v := x; return &v" boilerplate everywhere.
package ptr

// To returns a pointer to the given value. Useful for taking the address
// of a literal or a non-addressable expression, e.g. ptr.To("hello") or
// ptr.To(int32(1440)).
func To[T any](v T) *T {
	return &v
}

// Deref returns the value p points to, or fallback if p is nil.
func Deref[T any](p *T, fallback T) T {
	if p == nil {
		return fallback
	}
	return *p
}

// DerefOrZero returns the value p points to, or the zero value of T if
// p is nil.
func DerefOrZero[T any](p *T) T {
	var zero T
	return Deref(p, zero)
}

// Equal reports whether a and b point to equal values. Two nil pointers
// are considered equal; a nil paired with a non-nil pointer is not,
// regardless of the value the non-nil one points to.
func Equal[T comparable](a, b *T) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
