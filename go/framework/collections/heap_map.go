package collections

// Defines the interface for a Heap data structure that establishes the
// heap property based on the item values, but allows accessing items also through keys,
// which are not related to the heap property.
// This allows, e.g., updating items.
type HeapMap[K ~int | ~string, V any] interface {

	// Adds the specified item to the heap or updates an existing item, if the key is already present.
	AddOrReplace(key K, item V)

	// Returns the top-most item from the heap and removes it.
	// If there is no item in the heap, the second return value is false.
	Pop() (K, V, bool)

	// Returns the top-most item from the heap and removes it.
	// If there is no item in the heap, the second return value is false.
	Peek() (K, V, bool)

	// Returns the item with the specified key.
	GetByKey(key K) (V, bool)

	// Removes the item with the specified key from the heap and returns it.
	// If there is no item in the heap with that key, the second return value is false.
	RemoveByKey(key K) (V, bool)

	// Gets the number of items currently in the heap.
	Len() int
}
