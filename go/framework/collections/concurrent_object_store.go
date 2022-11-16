package collections

// Provides a thread-safe store for objects that can be accessed by key or index.
//
// This store is designed for batch access. Thus, its methods obtain a suitable lock and then
// supply a reader or a writer object, which can be used to make multiple reads or writes
// while keeping the lock in-between.
//
// Note that access through indices is only possible to allow random access
// and round-robin access to the stored objects. There is no particular guaranteed
// order to the objects retrieved by index.
// The only index-related property that is guaranteed is that the order stays the same
// as long as no object is added to or removed from the store.
type ConcurrentObjectStore[V any] interface {

	// Obtains a read lock on the store and returns a reader object.
	ReadLock() ConcurrentObjectStoreReader[V]

	// Obtains a write lock on the store and returns a writer object.
	WriteLock() ConcurrentObjectStoreWriter[V]
}

// Provides batch access to a ConcurrentObjectStore locked for reading.
type ConcurrentObjectStoreReader[V any] interface {
	// Gets the object with the specified key.
	GetByKey(key string) (V, bool)

	// Gets the key and the object stored at the specified index.
	//
	// If the index is out of bounds, the last return value is false.
	GetByIndex(index int) (string, V, bool)

	// Returns the number of objects stored.
	Len() int

	// Releases the lock. The store must not be accessed anymore after calling this method.
	Unlock()
}

// Provides batch access to a ConcurrentObjectStore locked for writing.
type ConcurrentObjectStoreWriter[V any] interface {
	ConcurrentObjectStoreReader[V]

	// Stores or updates the object with the specified key.
	Set(key string, value V)

	// Removes the object with the specified key.
	//
	// Returns the object that was stored or false, if the key did not exist.
	Remove(key string) (V, bool)
}
