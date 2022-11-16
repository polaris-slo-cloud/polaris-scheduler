package collections

import "sync"

var (
	_ ConcurrentObjectStore[any]       = (*ConcurrentObjectStoreImpl[any])(nil)
	_ ConcurrentObjectStoreReader[any] = (*concurrentObjectStoreImplReader[any])(nil)
	_ ConcurrentObjectStoreWriter[any] = (*concurrentObjectStoreImplWriter[any])(nil)
)

type keyValuePair[V any] struct {
	key   string
	value V
}

type concurrentObjectStoreImplReader[V any] struct {
	store *ConcurrentObjectStoreImpl[V]
}

type concurrentObjectStoreImplWriter[V any] struct {
	concurrentObjectStoreImplReader[V]
}

// Implementation of ConcurrentObjectStore.
type ConcurrentObjectStoreImpl[V any] struct {
	// Using a MinHeapMap and accessing its data store provides us with the following operation complexities:
	//
	// Get/Set by key O(1)
	// Get by index O(1)
	// Add O(1)
	// Remove O(log n)
	minHeap *MinHeapMap[string, *keyValuePair[V]]

	mutex *sync.RWMutex

	reader *concurrentObjectStoreImplReader[V]
	writer *concurrentObjectStoreImplWriter[V]

	nilValue V
}

// ConcurrentObjectStoreImpl

func NewConcurrentObjectStoreImpl[V any]() *ConcurrentObjectStoreImpl[V] {
	var nilValue V
	store := &ConcurrentObjectStoreImpl[V]{
		minHeap:  NewMinHeapMap[string, *keyValuePair[V]](keyValuePairLess[V], nil),
		mutex:    &sync.RWMutex{},
		nilValue: nilValue,
	}

	store.reader = &concurrentObjectStoreImplReader[V]{store: store}
	store.writer = &concurrentObjectStoreImplWriter[V]{
		concurrentObjectStoreImplReader: concurrentObjectStoreImplReader[V]{store: store},
	}

	return store
}

func (cos *ConcurrentObjectStoreImpl[V]) ReadLock() ConcurrentObjectStoreReader[V] {
	cos.mutex.RLock()
	return cos.reader
}

func (cos *ConcurrentObjectStoreImpl[V]) WriteLock() ConcurrentObjectStoreWriter[V] {
	cos.mutex.Lock()
	return cos.writer
}

func keyValuePairLess[V any](itemA *keyValuePair[V], itemB *keyValuePair[V]) bool {
	return itemA.key < itemB.key
}

// concurrentObjectStoreImplReader

func (r *concurrentObjectStoreImplReader[V]) GetByIndex(index int) (string, V, bool) {
	if index >= 0 && index < r.Len() {
		item := r.store.minHeap.data.items[index]
		return item.key, item.value.value, true
	}
	return "", r.store.nilValue, false
}

func (r *concurrentObjectStoreImplReader[V]) GetByKey(key string) (V, bool) {
	if item, ok := r.store.minHeap.GetByKey(key); ok {
		return item.value, true
	}
	return r.store.nilValue, false
}

func (r *concurrentObjectStoreImplReader[V]) Len() int {
	return r.store.minHeap.Len()
}

func (r *concurrentObjectStoreImplReader[V]) Unlock() {
	r.store.mutex.RUnlock()
}

// concurrentObjectStoreImplWriter

func (w *concurrentObjectStoreImplWriter[V]) Set(key string, value V) {
	w.store.minHeap.AddOrReplace(key, &keyValuePair[V]{key: key, value: value})
}

func (w *concurrentObjectStoreImplWriter[V]) Remove(key string) (V, bool) {
	if item, ok := w.store.minHeap.RemoveByKey(key); ok {
		return item.value, true
	}
	return w.store.nilValue, false
}

func (w *concurrentObjectStoreImplWriter[V]) Unlock() {
	w.store.mutex.Unlock()
}
