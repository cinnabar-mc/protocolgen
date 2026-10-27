// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

// Optional holds a value that may be absent from the wire.
type Optional[T any] struct {
	set bool
	val T
}

// Option creates a present Optional containing value.
func Option[T any](value T) Optional[T] {
	return Optional[T]{set: true, val: value}
}

// Value returns the optional value and whether it is present.
func (o Optional[T]) Value() (T, bool) {
	return o.val, o.set
}

// OrderedEntry preserves the source order and duplicate keys of a wire map.
type OrderedEntry[K, V any] struct {
	Key   K
	Value V
}
