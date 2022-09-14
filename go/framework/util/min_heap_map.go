package util

import (
	"container/heap"
)

var (
	_ HeapMap[string, any] = (*MinHeapMap[string, any])(nil)
)

// Implements a HeapMap with the "smallest" item being on top of the heap.
type MinHeapMap[K ~int | ~string, V any] struct {
	data     *heapMapData[K, V]
	nilKey   K
	nilValue V
}

// Creates a new MinHeap with the specified function LessFunc for comparing items to establish the min heap property.
//
// The nilValue is required, because even though `any` is defined as interface{}, Go allows also non-pointer types
// types to be used for such generic parameters.
func NewMinHeapMap[K ~int | ~string, V any](lessFn LessFunc[V], nilValue V) *MinHeapMap[K, V] {
	h := MinHeapMap[K, V]{
		data:     newHeapData[K, V](lessFn),
		nilKey:   *new(K),
		nilValue: nilValue,
	}
	heap.Init(h.data)
	return &h
}

func (mh *MinHeapMap[K, V]) AddOrReplace(key K, item V) {
	if mh.data.exists(key) {
		mh.data.updateItem(key, item)
	} else {
		item := mh.data.newHeapItem(key, item)
		heap.Push(mh.data, item)
	}
}

func (mh *MinHeapMap[K, V]) Len() int {
	return mh.data.Len()
}

func (mh *MinHeapMap[K, V]) Peek() (K, V, bool) {
	item, ok := mh.data.peek()
	if ok {
		return item.key, item.value, true
	}
	return mh.nilKey, mh.nilValue, false
}

func (mh *MinHeapMap[K, V]) Pop() (K, V, bool) {
	if mh.Len() == 0 {
		return mh.nilKey, mh.nilValue, false
	}
	item := heap.Pop(mh.data).(*heapMapItem[K, V])
	return item.key, item.value, true
}

func (mh *MinHeapMap[K, V]) GetByKey(key K) (V, bool) {
	item, ok := mh.data.getByKey(key)
	if ok {
		return item.value, true
	}
	return mh.nilValue, false
}

func (mh *MinHeapMap[K, V]) RemoveByKey(key K) (V, bool) {
	item, ok := mh.data.getByKey(key)
	if !ok {
		return mh.nilValue, false
	}
	heap.Remove(mh.data, item.index)
	return item.value, true
}
