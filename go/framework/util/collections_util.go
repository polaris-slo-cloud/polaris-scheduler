package util

// Defines a comparison function that returns true iff itemA is less than itemB, otherwise false.
type LessFunc[T any] func(itemA T, itemB T) bool

// Returns the key for the specified value (can be used for maps or similar data structures).
type KeyFunc[K ~int | ~string, V any] func(value V) (K, error)
