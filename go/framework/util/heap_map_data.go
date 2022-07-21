package util

import (
	"container/heap"
)

var (
	_ heap.Interface = (*heapMapData[string, any])(nil)
)

// Represents an item that is stored in the heap.
type heapMapItem[K ~int | ~string, V any] struct {
	// The key used to identify the item.
	key K

	// The actual item value.
	value V

	// The index in the heap's items array, where the item is currently stored.
	// IMPORTANT: This value should only be modified by heapMapData methods.
	index int
}

// Stores the data for a heap.
// The heap property is established based on the V values, but each item is also accessible through a key K.
//
// The heap property is established by comparing the item values using a the LessFunc.
// Additionally, each item is also accessible through a key, which does not influence the heap property.
type heapMapData[K ~int | ~string, V any] struct {

	// Stores the items with the maintained heap property.
	items []*heapMapItem[K, V]

	// Stores the items by their key to allow looking up items in the heap by key.
	itemsByKey map[K]*heapMapItem[K, V]

	// Used to extract the key from a value.
	keyFn KeyFunc[K, V]

	// Used to determine the precedence between two items for establishing the heap property
	lessFn LessFunc[V]
}

// ToDo: implement shrinking the items slice if its capacity is way beyond the current length.

// Creates a new heapData with the specified LessFunc for comparing items to establish the min heap property.
func newHeapData[K ~int | ~string, V any](lessFn LessFunc[V]) *heapMapData[K, V] {
	h := heapMapData[K, V]{
		items:      make([]*heapMapItem[K, V], 0),
		itemsByKey: make(map[K]*heapMapItem[K, V], 0),
		lessFn:     lessFn,
	}
	return &h
}

// Creates a new heapItem for adding to this heap.
// IMPORTANT: This method does not add anything to the heap.
func (h *heapMapData[K, V]) newHeapItem(key K, value V) *heapMapItem[K, V] {
	return &heapMapItem[K, V]{
		key:   key,
		value: value,
		index: -1,
	}
}

func (h *heapMapData[K, V]) Len() int {
	return len(h.items)
}

func (h *heapMapData[K, V]) Less(i int, j int) bool {
	itemI := h.items[i]
	itemJ := h.items[j]
	return h.lessFn(itemI.value, itemJ.value)
}

func (h *heapMapData[K, V]) Swap(i int, j int) {
	itemI := h.items[i]
	itemJ := h.items[j]
	h.items[i] = itemJ
	h.items[j] = itemI
	itemJ.index = i
	itemI.index = j
}

func (h *heapMapData[K, V]) Push(x any) {
	item, ok := x.(*heapMapItem[K, V])
	if !ok {
		panic("This heap only supports items of the type it was instantiated with.")
	}

	item.index = len(h.items)
	h.items = append(h.items, item)
	h.itemsByKey[item.key] = item
}

func (h *heapMapData[K, V]) Pop() any {
	n := len(h.items)
	lastIndex := n - 1
	if lastIndex == -1 {
		return nil
	}
	item := h.items[lastIndex]

	item.index = -1
	h.items[lastIndex] = nil // Avoid memory leak.
	h.items = h.items[0:lastIndex]
	delete(h.itemsByKey, item.key)

	return item
}

// Gets an item by its key.
func (h *heapMapData[K, V]) getByKey(key K) (*heapMapItem[K, V], bool) {
	item, ok := h.itemsByKey[key]
	if ok {
		return item, ok
	}
	return nil, false
}

// Determines if an item with the given key exists in the heap.
func (h *heapMapData[K, V]) exists(key K) bool {
	_, ok := h.itemsByKey[key]
	return ok
}

// Updates the item in the heap and then reestablishes the heap property.
// Returns true if the update was successful or false if no item with the keyFn(x) exists in the heap.
func (h *heapMapData[K, V]) updateItem(key K, value V) bool {
	item, ok := h.itemsByKey[key]
	if ok {
		item.value = value
		heap.Fix(h, item.index)
	}
	return ok
}

// Returns the top-most item without removing it.
func (h *heapMapData[K, V]) peek() (*heapMapItem[K, V], bool) {
	if len(h.items) == 0 {
		return nil, false
	}

	return h.items[0], true
}
